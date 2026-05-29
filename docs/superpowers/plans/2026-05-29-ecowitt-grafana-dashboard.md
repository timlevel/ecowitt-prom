# Ecowitt Grafana Dashboard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create a Grafana dashboard with canvas summary panels and time series trends, split into Outdoor and Indoor rows.

**Architecture:** A Python build script generates the full dashboard JSON and POSTs it to the Grafana HTTP API. Canvas panels use instant Prometheus queries with reduce+merge transformations so each element can bind to a named field. Time series panels use range queries. Output JSON is saved to `grafana/ecowitt-dashboard.json` for reproducibility.

**Tech Stack:** Python 3, requests, Grafana HTTP API (`/api/dashboards/db`), Prometheus datasource UID `prometheus`, Grafana at `http://localhost:3001` (admin/admin).

---

### Task 1: Create the dashboard build script

**Files:**
- Create: `grafana/build_dashboard.py`

- [ ] **Step 1: Create `grafana/` directory and build script**

```python
#!/usr/bin/env python3
"""Build and POST the ecowitt Grafana dashboard."""
import json
import requests

GRAFANA_URL = "http://localhost:3001"
GRAFANA_AUTH = ("admin", "admin")
DS = {"type": "prometheus", "uid": "prometheus"}


# ── Helpers ──────────────────────────────────────────────────────────────────

def target(expr, legend, ref_id, instant=False):
    return {
        "datasource": DS,
        "expr": expr,
        "legendFormat": legend,
        "instant": instant,
        "range": not instant,
        "refId": ref_id,
        "editorMode": "code",
    }


def row_panel(pid, title, y):
    return {
        "id": pid, "type": "row", "title": title,
        "gridPos": {"h": 1, "w": 24, "x": 0, "y": y},
        "collapsed": False, "panels": [],
    }


def ts_panel(pid, title, x, y, targets, unit="", w=12):
    return {
        "id": pid, "type": "timeseries", "title": title,
        "gridPos": {"h": 8, "w": w, "x": x, "y": y},
        "datasource": DS,
        "targets": targets,
        "fieldConfig": {
            "defaults": {
                "unit": unit,
                "color": {"mode": "palette-classic"},
                "custom": {"lineWidth": 2, "fillOpacity": 10},
            },
        },
        "options": {
            "tooltip": {"mode": "multi", "sort": "none"},
            "legend": {"displayMode": "list", "placement": "bottom"},
        },
    }


def label_el(name, text, lx, ty, w, h, size=11, color="#a0a0a0"):
    return {
        "config": {
            "align": "center", "valign": "middle",
            "color": {"fixed": color},
            "size": size,
            "text": {"fixed": text},
        },
        "connections": [], "data": {"source": "fixed"},
        "name": name,
        "placement": {"left": lx, "top": ty, "width": w, "height": h},
        "type": "text",
    }


def value_el(name, field, lx, ty, w, h, size=32, color="#ffffff", unit=""):
    return {
        "config": {
            "align": "center", "valign": "middle",
            "color": {"fixed": color},
            "size": size,
            "text": {"fixed": ""},
            "unit": unit,
        },
        "connections": [], "data": {"field": field, "source": "data"},
        "name": name,
        "placement": {"left": lx, "top": ty, "width": w, "height": h},
        "type": "metric-value",
    }


def canvas_panel(pid, title, y, targets, elements):
    return {
        "id": pid, "type": "canvas", "title": title,
        "gridPos": {"h": 14, "w": 24, "x": 0, "y": y},
        "datasource": DS,
        "targets": targets,
        "transformations": [
            {"id": "reduce", "options": {"reducers": ["lastNotNull"]}},
            {"id": "merge", "options": {}},
        ],
        "options": {
            "inlineEditing": False,
            "panZoom": False,
            "showAdvancedTypes": True,
            "root": {
                "border": "Dark Lines",
                "layout": {"kind": "free-layout"},
                "name": title,
                "type": "frame",
                "elements": elements,
            },
        },
    }


# ── Outdoor canvas ────────────────────────────────────────────────────────────

outdoor_canvas_targets = [
    target('ecowitt_temperature_celsius{location="outdoor"}', "out_temp",   "A", instant=True),
    target('ecowitt_humidity_percent{location="outdoor"}',    "out_hum",    "B", instant=True),
    target("ecowitt_wind_speed_kph",                          "wind_speed", "C", instant=True),
    target("ecowitt_wind_gust_kph",                           "wind_gust",  "D", instant=True),
    target("ecowitt_wind_direction_degrees",                  "wind_dir",   "E", instant=True),
    target('ecowitt_pressure_hpa{type="relative"}',           "pres_rel",   "F", instant=True),
    target('ecowitt_pressure_hpa{type="absolute"}',           "pres_abs",   "G", instant=True),
    target("ecowitt_solar_radiation_wm2",                     "solar",      "H", instant=True),
    target("ecowitt_uv_index",                                "uv",         "I", instant=True),
    target("ecowitt_rain_rate_mm_per_hour",                   "rain_rate",  "J", instant=True),
]

outdoor_canvas_elements = [
    # Temperature block (left)
    label_el("lbl_otemp", "OUTDOOR TEMP",  20,  8, 160, 22, size=11),
    value_el("val_otemp", "out_temp",      20, 32, 160, 65, size=42, color="#74c1e8", unit="°C"),

    # Humidity
    label_el("lbl_ohum", "HUMIDITY",      200,  8, 110, 22, size=11),
    value_el("val_ohum", "out_hum",       200, 32, 110, 65, size=36, color="#74c8c8", unit="%"),

    # Wind block (centre-right)
    label_el("lbl_wspd", "WIND",          350,  8, 110, 22, size=11),
    value_el("val_wspd", "wind_speed",    350, 32, 110, 65, size=34, color="#e8d074", unit=" km/h"),

    label_el("lbl_wgst", "GUST",          480,  8, 100, 22, size=11),
    value_el("val_wgst", "wind_gust",     480, 32, 100, 65, size=30, color="#e8b074", unit=" km/h"),

    label_el("lbl_wdir", "DIRECTION",     600,  8,  90, 22, size=11),
    value_el("val_wdir", "wind_dir",      600, 32,  90, 65, size=28, color="#c8c8c8", unit="°"),

    # Pressure (second row)
    label_el("lbl_prel", "REL PRESSURE",  20, 115, 150, 22, size=11),
    value_el("val_prel", "pres_rel",      20, 140, 150, 55, size=28, color="#c8e8a0", unit=" hPa"),

    label_el("lbl_pabs", "ABS PRESSURE", 190, 115, 150, 22, size=11),
    value_el("val_pabs", "pres_abs",     190, 140, 150, 55, size=28, color="#a0c8a0", unit=" hPa"),

    # Solar / UV (second row right)
    label_el("lbl_sol", "SOLAR",          350, 115, 110, 22, size=11),
    value_el("val_sol", "solar",          350, 140, 110, 55, size=28, color="#f5d76e", unit=" W/m²"),

    label_el("lbl_uv",  "UV INDEX",       480, 115,  90, 22, size=11),
    value_el("val_uv",  "uv",             480, 140,  90, 55, size=32, color="#f5a623"),

    # Rain rate (bottom)
    label_el("lbl_rain", "RAIN RATE",      20, 215, 130, 22, size=11),
    value_el("val_rain", "rain_rate",      20, 240, 130, 45, size=24, color="#74a8e8", unit=" mm/h"),
]

outdoor_canvas = canvas_panel(
    pid=2, title="Outdoor Conditions",
    y=1,
    targets=outdoor_canvas_targets,
    elements=outdoor_canvas_elements,
)


# ── Indoor canvas ─────────────────────────────────────────────────────────────

indoor_canvas_targets = [
    target('ecowitt_temperature_celsius{location="indoor"}',  "in_temp",  "A", instant=True),
    target('ecowitt_humidity_percent{location="indoor"}',     "in_hum",   "B", instant=True),
    target('ecowitt_temperature_celsius{location="ch1"}',     "ch1_temp", "C", instant=True),
    target('ecowitt_humidity_percent{location="ch1"}',        "ch1_hum",  "D", instant=True),
    target('ecowitt_temperature_celsius{location="ch2"}',     "ch2_temp", "E", instant=True),
    target('ecowitt_humidity_percent{location="ch2"}',        "ch2_hum",  "F", instant=True),
    target('ecowitt_temperature_celsius{location="ch3"}',     "ch3_temp", "G", instant=True),
    target('ecowitt_humidity_percent{location="ch3"}',        "ch3_hum",  "H", instant=True),
    target('ecowitt_battery{sensor="wh65batt"}',              "batt_wh65","I", instant=True),
    target('ecowitt_battery{sensor="ch1"}',                   "batt_ch1", "J", instant=True),
    target('ecowitt_battery{sensor="ch2"}',                   "batt_ch2", "K", instant=True),
    target('ecowitt_battery{sensor="ch3"}',                   "batt_ch3", "L", instant=True),
]

indoor_canvas_elements = [
    # Indoor block (left)
    label_el("lbl_itemp", "INDOOR TEMP",  20,  8, 150, 22, size=11),
    value_el("val_itemp", "in_temp",      20, 32, 150, 65, size=42, color="#e87474", unit="°C"),

    label_el("lbl_ihum",  "HUMIDITY",    190,  8, 110, 22, size=11),
    value_el("val_ihum",  "in_hum",      190, 32, 110, 65, size=36, color="#e8a074", unit="%"),

    # Channel sensors grid (right)
    label_el("lbl_ch",    "CHANNEL SENSORS", 360, 8, 300, 22, size=11),

    label_el("lbl_ch1t",  "Ch1 Temp",   360, 35,  80, 20, size=10),
    value_el("val_ch1t",  "ch1_temp",   360, 57,  80, 45, size=22, color="#e8c8a0", unit="°C"),

    label_el("lbl_ch1h",  "Ch1 Hum",    460, 35,  80, 20, size=10),
    value_el("val_ch1h",  "ch1_hum",    460, 57,  80, 45, size=22, color="#a0c8e8", unit="%"),

    label_el("lbl_ch2t",  "Ch2 Temp",   360, 112,  80, 20, size=10),
    value_el("val_ch2t",  "ch2_temp",   360, 134,  80, 45, size=22, color="#e8c8a0", unit="°C"),

    label_el("lbl_ch2h",  "Ch2 Hum",    460, 112,  80, 20, size=10),
    value_el("val_ch2h",  "ch2_hum",    460, 134,  80, 45, size=22, color="#a0c8e8", unit="%"),

    label_el("lbl_ch3t",  "Ch3 Temp",   360, 189,  80, 20, size=10),
    value_el("val_ch3t",  "ch3_temp",   360, 211,  80, 45, size=22, color="#e8c8a0", unit="°C"),

    label_el("lbl_ch3h",  "Ch3 Hum",    460, 189,  80, 20, size=10),
    value_el("val_ch3h",  "ch3_hum",    460, 211,  80, 45, size=22, color="#a0c8e8", unit="%"),

    # Battery row (bottom)
    label_el("lbl_batt",  "BATTERIES",   20, 120, 300, 22, size=11),

    label_el("lbl_bwh",  "WH65",    20, 148,  60, 18, size=10),
    value_el("val_bwh",  "batt_wh65",20, 168,  60, 35, size=18, color="#a0e8a0"),

    label_el("lbl_bc1",  "Ch1",     100, 148,  50, 18, size=10),
    value_el("val_bc1",  "batt_ch1",100, 168,  50, 35, size=18, color="#a0e8a0"),

    label_el("lbl_bc2",  "Ch2",     165, 148,  50, 18, size=10),
    value_el("val_bc2",  "batt_ch2",165, 168,  50, 35, size=18, color="#a0e8a0"),

    label_el("lbl_bc3",  "Ch3",     230, 148,  50, 18, size=10),
    value_el("val_bc3",  "batt_ch3",230, 168,  50, 35, size=18, color="#a0e8a0"),
]

indoor_canvas = canvas_panel(
    pid=10, title="Indoor Conditions",
    y=40,
    targets=indoor_canvas_targets,
    elements=indoor_canvas_elements,
)


# ── Outdoor time series ───────────────────────────────────────────────────────

outdoor_ts = [
    ts_panel(3, "Outdoor Temperature", 0, 15, [
        target('ecowitt_temperature_celsius{location="outdoor"}', "Outdoor", "A"),
    ], unit="celsius"),

    ts_panel(4, "Outdoor Humidity", 12, 15, [
        target('ecowitt_humidity_percent{location="outdoor"}', "Outdoor", "A"),
    ], unit="percent"),

    ts_panel(5, "Barometric Pressure", 0, 23, [
        target('ecowitt_pressure_hpa{type="relative"}', "Relative", "A"),
        target('ecowitt_pressure_hpa{type="absolute"}', "Absolute", "B"),
    ], unit="pressurehpa"),

    ts_panel(6, "Wind Speed & Gust", 12, 23, [
        target("ecowitt_wind_speed_kph",      "Speed", "A"),
        target("ecowitt_wind_gust_kph",       "Gust",  "B"),
        target("ecowitt_max_daily_gust_kph",  "Max Daily Gust", "C"),
    ], unit="velocitykmh"),

    ts_panel(7, "Rain Accumulation", 0, 31, [
        target('ecowitt_rain_mm{period="hourly"}',  "Hourly",  "A"),
        target('ecowitt_rain_mm{period="daily"}',   "Daily",   "B"),
        target('ecowitt_rain_mm{period="weekly"}',  "Weekly",  "C"),
        target('ecowitt_rain_mm{period="monthly"}', "Monthly", "D"),
    ], unit="lengthmm"),

    ts_panel(8, "Solar Radiation & UV", 12, 31, [
        target("ecowitt_solar_radiation_wm2", "Solar (W/m²)", "A"),
        target("ecowitt_uv_index",            "UV Index",     "B"),
    ]),
]


# ── Indoor time series ────────────────────────────────────────────────────────

indoor_ts = [
    ts_panel(11, "Indoor Temperature", 0, 54, [
        target('ecowitt_temperature_celsius{location="indoor"}', "Indoor", "A"),
    ], unit="celsius"),

    ts_panel(12, "Indoor Humidity", 12, 54, [
        target('ecowitt_humidity_percent{location="indoor"}', "Indoor", "A"),
    ], unit="percent"),

    ts_panel(13, "Channel Temperatures", 0, 62, [
        target('ecowitt_temperature_celsius{location="ch1"}', "Ch1", "A"),
        target('ecowitt_temperature_celsius{location="ch2"}', "Ch2", "B"),
        target('ecowitt_temperature_celsius{location="ch3"}', "Ch3", "C"),
    ], unit="celsius"),

    ts_panel(14, "Channel Humidity", 12, 62, [
        target('ecowitt_humidity_percent{location="ch1"}', "Ch1", "A"),
        target('ecowitt_humidity_percent{location="ch2"}', "Ch2", "B"),
        target('ecowitt_humidity_percent{location="ch3"}', "Ch3", "C"),
    ], unit="percent"),
]


# ── Assemble dashboard ────────────────────────────────────────────────────────

dashboard = {
    "id": None,
    "uid": "ecowitt-weather",
    "title": "Ecowitt Weather Station",
    "tags": ["ecowitt", "weather"],
    "timezone": "browser",
    "refresh": "60s",
    "time": {"from": "now-24h", "to": "now"},
    "schemaVersion": 39,
    "panels": [
        row_panel(1, "🌤  Outdoor", y=0),
        outdoor_canvas,
        *outdoor_ts,
        row_panel(9, "🏠  Indoor", y=39),
        indoor_canvas,
        *indoor_ts,
    ],
}

payload = {"dashboard": dashboard, "overwrite": True, "folderId": 0}


# ── POST to Grafana ───────────────────────────────────────────────────────────

resp = requests.post(
    f"{GRAFANA_URL}/api/dashboards/db",
    auth=GRAFANA_AUTH,
    json=payload,
    headers={"Content-Type": "application/json"},
)
resp.raise_for_status()
result = resp.json()
print(f"Dashboard URL: {GRAFANA_URL}{result['url']}")

# Save JSON for reproducibility
with open("grafana/ecowitt-dashboard.json", "w") as f:
    json.dump(payload, f, indent=2)
print("Saved to grafana/ecowitt-dashboard.json")
```

- [ ] **Step 2: Run the script**

```bash
cd /home/timlevett/projects/ecowitt-prom
pip install requests -q
python grafana/build_dashboard.py
```

Expected output:
```
Dashboard URL: http://localhost:3001/d/ecowitt-weather/ecowitt-weather-station
Saved to grafana/ecowitt-dashboard.json
```

- [ ] **Step 3: Verify dashboard exists in Grafana**

```bash
curl -s http://admin:admin@localhost:3001/api/dashboards/uid/ecowitt-weather | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['dashboard']['title'], '—', d['meta']['url'])"
```

Expected: `Ecowitt Weather Station — /d/ecowitt-weather/ecowitt-weather-station`

- [ ] **Step 4: Commit**

```bash
git add grafana/build_dashboard.py grafana/ecowitt-dashboard.json
git commit -m "feat: add Grafana dashboard for ecowitt weather station"
```
