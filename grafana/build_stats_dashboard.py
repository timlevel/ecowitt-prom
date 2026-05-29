#!/usr/bin/env python3
"""Build and POST the ecowitt stats+trends Grafana dashboard (Grafana 13 tabs)."""
import json
import requests

GRAFANA_URL = "http://localhost:3001"
GRAFANA_AUTH = ("admin", "admin")
DS = {"type": "prometheus", "uid": "prometheus"}


# ── Query helpers ─────────────────────────────────────────────────────────────

def t(expr, legend, ref_id):
    return {"datasource": DS, "expr": expr, "legendFormat": legend,
            "refId": ref_id, "editorMode": "code"}

def temp_f(selector):
    """Convert Celsius metric to Fahrenheit in PromQL."""
    return f"({selector} * 9/5) + 32"


# ── Threshold helpers ─────────────────────────────────────────────────────────

def steps(*pairs):
    """Build threshold steps: steps(("green", None), ("yellow", 30), ...)"""
    return [{"color": c, "value": v} for c, v in pairs]

THRESH = {
    "temp_f": steps(
        ("blue",       None),   # < 32°F  freezing
        ("light-blue", 32),     # 32–50   cold
        ("super-light-blue", 50),# 50–68  cool
        ("green",      68),     # 68–86   comfortable
        ("yellow",     86),     # 86–95   warm
        ("red",        95),     # 95+     hot
    ),
    "humidity": steps(
        ("orange",     None),   # < 30%  dry
        ("green",      30),     # 30–60  comfortable
        ("yellow",     60),     # 60–80  humid
        ("blue",       80),     # 80+    very humid
    ),
    "wind_kmh": steps(
        ("green",      None),   # < 20   calm
        ("yellow",     20),     # 20–40  moderate
        ("orange",     40),     # 40–60  fresh
        ("red",        60),     # 60+    strong
    ),
    "uv": steps(
        ("green",      None),   # 0–3    low
        ("yellow",     3),      # 3–6    moderate
        ("orange",     6),      # 6–8    high
        ("red",        8),      # 8–11   very high
        ("purple",     11),     # 11+    extreme
    ),
    "pressure": steps(
        ("orange",     None),   # < 1000  stormy
        ("green",      1000),   # 1000–1020  normal
        ("blue",       1020),   # 1020+  fair/high
    ),
    "solar": steps(
        ("text",       None),   # 0      no sun
        ("yellow",     1),      # 1–200  low
        ("orange",     200),    # 200–500 moderate
        ("red",        500),    # 500+   bright
    ),
    "rain_rate": steps(
        ("green",      None),   # 0      dry
        ("light-blue", 0.1),    # 0.1–2.5 light
        ("blue",       2.5),    # 2.5–10  moderate
        ("dark-blue",  10),     # 10+    heavy
    ),
    "battery": steps(
        ("red",        None),   # 0  dead/low
        ("green",      1),      # 1+ good
    ),
    "soil_moisture": steps(
        ("orange",     None),   # < 20%  dry
        ("green",      20),     # 20–60  adequate
        ("blue",       60),     # 60+    wet
    ),
}


# ── Panel builders ────────────────────────────────────────────────────────────

_pid = [100]

def next_id():
    _pid[0] += 1
    return _pid[0]


def stat(title, x, y, targets, unit, thresh_key, w=6, h=5):
    return {
        "id": next_id(), "type": "stat", "title": title,
        "gridPos": {"h": h, "w": w, "x": x, "y": y},
        "datasource": DS,
        "targets": targets,
        "options": {
            "reduceOptions": {"values": False, "calcs": ["lastNotNull"], "fields": ""},
            "orientation": "auto",
            "textMode": "auto",
            "colorMode": "background",
            "graphMode": "area",
            "justifyMode": "auto",
        },
        "fieldConfig": {
            "defaults": {
                "unit": unit,
                "color": {"mode": "thresholds"},
                "thresholds": {"mode": "absolute", "steps": THRESH[thresh_key]},
                "custom": {"fillOpacity": 10},
            },
        },
    }


def ts(title, x, y, targets, unit, w=12, h=8):
    return {
        "id": next_id(), "type": "timeseries", "title": title,
        "gridPos": {"h": h, "w": w, "x": x, "y": y},
        "datasource": DS,
        "targets": targets,
        "fieldConfig": {
            "defaults": {
                "unit": unit,
                "color": {"mode": "palette-classic"},
                "custom": {"lineWidth": 2, "fillOpacity": 8},
            },
        },
        "options": {
            "tooltip": {"mode": "multi", "sort": "none"},
            "legend": {"displayMode": "list", "placement": "bottom"},
        },
    }


def row_sep(title, y=0):
    return {
        "id": next_id(), "type": "row", "title": title,
        "gridPos": {"h": 1, "w": 24, "x": 0, "y": y},
        "collapsed": False,
        "panels": [],
    }


def offset_y(panels, dy):
    result = []
    for p in panels:
        p2 = dict(p)
        gp = dict(p["gridPos"])
        gp["y"] = gp["y"] + dy
        p2["gridPos"] = gp
        result.append(p2)
    return result


# ── Outdoor tab ───────────────────────────────────────────────────────────────

outdoor_panels = [
    # ── Stat row 1 (y=0): temperature, humidity, wind speed, gust ──
    stat("Outdoor Temp", 0, 0,
         [t(temp_f('ecowitt_temperature_celsius{location="outdoor"}'), "Outdoor", "A")],
         "fahrenheit", "temp_f"),

    stat("Outdoor Humidity", 6, 0,
         [t('ecowitt_humidity_percent{location="outdoor"}', "Outdoor", "A")],
         "percent", "humidity"),

    stat("Wind Speed", 12, 0,
         [t("ecowitt_wind_speed_kph", "Speed", "A")],
         "velocitykmh", "wind_kmh"),

    stat("Wind Gust", 18, 0,
         [t("ecowitt_wind_gust_kph", "Gust", "A")],
         "velocitykmh", "wind_kmh"),

    # ── Stat row 2 (y=5): pressure, UV, solar, rain rate ──
    stat("Rel. Pressure", 0, 5,
         [t('ecowitt_pressure_hpa{type="relative"}', "Relative", "A")],
         "pressurehpa", "pressure"),

    stat("UV Index", 6, 5,
         [t("ecowitt_uv_index", "UV", "A")],
         "none", "uv"),

    stat("Solar Radiation", 12, 5,
         [t("ecowitt_solar_radiation_wm2", "Solar", "A")],
         "watt_per_meterM2", "solar"),

    stat("Rain Rate", 18, 5,
         [t("ecowitt_rain_rate_mm_per_hour", "Rate", "A")],
         "lengthmm", "rain_rate"),

    # ── Time series row 1 (y=10) ──
    ts("Temperature", 0, 10,
       [t(temp_f('ecowitt_temperature_celsius{location="outdoor"}'), "Outdoor °F", "A")],
       "fahrenheit"),

    ts("Humidity", 12, 10,
       [t('ecowitt_humidity_percent{location="outdoor"}', "Outdoor", "A")],
       "percent"),

    # ── Time series row 2 (y=18) ──
    ts("Barometric Pressure", 0, 18,
       [t('ecowitt_pressure_hpa{type="relative"}', "Relative", "A"),
        t('ecowitt_pressure_hpa{type="absolute"}', "Absolute", "B")],
       "pressurehpa"),

    ts("Wind Speed & Gust", 12, 18,
       [t("ecowitt_wind_speed_kph",     "Speed",          "A"),
        t("ecowitt_wind_gust_kph",      "Gust",           "B"),
        t("ecowitt_max_daily_gust_kph", "Max Daily Gust", "C")],
       "velocitykmh"),

    # ── Time series row 3 (y=26) ──
    ts("Rain Accumulation", 0, 26,
       [t('ecowitt_rain_mm{period="hourly"}',  "Hourly",  "A"),
        t('ecowitt_rain_mm{period="daily"}',   "Daily",   "B"),
        t('ecowitt_rain_mm{period="weekly"}',  "Weekly",  "C"),
        t('ecowitt_rain_mm{period="monthly"}', "Monthly", "D")],
       "lengthmm"),

    ts("Solar Radiation & UV", 12, 26,
       [t("ecowitt_solar_radiation_wm2", "Solar (W/m²)", "A"),
        t("ecowitt_uv_index",            "UV Index",     "B")],
       "none"),
]


# ── Indoor tab ────────────────────────────────────────────────────────────────

indoor_panels = [
    # ── Stat row 1 (y=0): indoor + ch1 + ch2 ──
    stat("Indoor Temp", 0, 0,
         [t(temp_f('ecowitt_temperature_celsius{location="indoor"}'), "Indoor", "A")],
         "fahrenheit", "temp_f", w=4),

    stat("Indoor Humidity", 4, 0,
         [t('ecowitt_humidity_percent{location="indoor"}', "Indoor", "A")],
         "percent", "humidity", w=4),

    stat("Ch1 Temp", 8, 0,
         [t(temp_f('ecowitt_temperature_celsius{location="ch1"}'), "Ch1", "A")],
         "fahrenheit", "temp_f", w=4),

    stat("Ch1 Humidity", 12, 0,
         [t('ecowitt_humidity_percent{location="ch1"}', "Ch1", "A")],
         "percent", "humidity", w=4),

    stat("Ch2 Temp", 16, 0,
         [t(temp_f('ecowitt_temperature_celsius{location="ch2"}'), "Ch2", "A")],
         "fahrenheit", "temp_f", w=4),

    stat("Ch2 Humidity", 20, 0,
         [t('ecowitt_humidity_percent{location="ch2"}', "Ch2", "A")],
         "percent", "humidity", w=4),

    # ── Stat row 2 (y=5): ch3 + batteries ──
    stat("Ch3 Temp", 0, 5,
         [t(temp_f('ecowitt_temperature_celsius{location="ch3"}'), "Ch3", "A")],
         "fahrenheit", "temp_f", w=4),

    stat("Ch3 Humidity", 4, 5,
         [t('ecowitt_humidity_percent{location="ch3"}', "Ch3", "A")],
         "percent", "humidity", w=4),

    stat("Batt WH65", 8, 5,
         [t('ecowitt_battery{sensor="wh65batt"}', "WH65", "A")],
         "none", "battery", w=4),

    stat("Batt Ch1", 12, 5,
         [t('ecowitt_battery{sensor="ch1"}', "Ch1", "A")],
         "none", "battery", w=4),

    stat("Batt Ch2", 16, 5,
         [t('ecowitt_battery{sensor="ch2"}', "Ch2", "A")],
         "none", "battery", w=4),

    stat("Batt Ch3", 20, 5,
         [t('ecowitt_battery{sensor="ch3"}', "Ch3", "A")],
         "none", "battery", w=4),

    # ── Time series (y=10) ──
    ts("Indoor Temperature", 0, 10,
       [t(temp_f('ecowitt_temperature_celsius{location="indoor"}'), "Indoor °F", "A")],
       "fahrenheit"),

    ts("Indoor Humidity", 12, 10,
       [t('ecowitt_humidity_percent{location="indoor"}', "Indoor", "A")],
       "percent"),

    # ── Time series (y=18) ──
    ts("Channel Temperatures", 0, 18,
       [t(temp_f('ecowitt_temperature_celsius{location="ch1"}'), "Ch1 °F", "A"),
        t(temp_f('ecowitt_temperature_celsius{location="ch2"}'), "Ch2 °F", "B"),
        t(temp_f('ecowitt_temperature_celsius{location="ch3"}'), "Ch3 °F", "C")],
       "fahrenheit"),

    ts("Channel Humidity", 12, 18,
       [t('ecowitt_humidity_percent{location="ch1"}', "Ch1", "A"),
        t('ecowitt_humidity_percent{location="ch2"}', "Ch2", "B"),
        t('ecowitt_humidity_percent{location="ch3"}', "Ch3", "C")],
       "percent"),
]


# ── Assemble & POST ───────────────────────────────────────────────────────────

dashboard = {
    "id": None,
    "uid": "ecowitt-stats",
    "title": "Ecowitt Weather Station — Stats",
    "tags": ["ecowitt", "weather"],
    "timezone": "browser",
    "refresh": "60s",
    "time": {"from": "now-24h", "to": "now"},
    "schemaVersion": 42,
    "panels": [
        row_sep("🌤  Outdoor", y=0),
        *offset_y(outdoor_panels, 1),
        row_sep("🏠  Indoor", y=35),
        *offset_y(indoor_panels, 36),
    ],
}

payload = {"dashboard": dashboard, "overwrite": True, "folderId": 0}

resp = requests.post(
    f"{GRAFANA_URL}/api/dashboards/db",
    auth=GRAFANA_AUTH,
    json=payload,
    headers={"Content-Type": "application/json"},
)
resp.raise_for_status()
result = resp.json()
print(f"Dashboard URL: {GRAFANA_URL}{result['url']}")

with open("grafana/ecowitt-stats-dashboard.json", "w") as f:
    json.dump(payload, f, indent=2)
print("Saved to grafana/ecowitt-stats-dashboard.json")
