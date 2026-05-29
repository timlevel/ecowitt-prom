# Ecowitt Grafana Dashboard Design

## Overview

A Grafana dashboard for the ecowitt-prom Prometheus exporter, visualising data from a WS3900B weather station. Two tabs — Outdoor and Indoor — each with a canvas summary panel at the top and time series trend panels below. Auto-refreshes every 60s to match station upload interval. Default time range: last 24h.

Dashboard is created via the Grafana HTTP API and stored as a JSON provisioning file in this repo.

---

## Data Source

Prometheus datasource within the `grafana/otel-lgtm` stack (default UID: `prometheus`). All metrics use the `ecowitt_` namespace with `station_type` and `model` labels from the WS3900B.

**Available metrics:**

| Metric | Labels |
|---|---|
| `ecowitt_temperature_celsius` | location: outdoor, indoor, ch1–ch3 |
| `ecowitt_humidity_percent` | location: outdoor, indoor, ch1–ch3 |
| `ecowitt_pressure_hpa` | type: relative, absolute |
| `ecowitt_wind_speed_kph` | — |
| `ecowitt_wind_gust_kph` | — |
| `ecowitt_max_daily_gust_kph` | — |
| `ecowitt_wind_direction_degrees` | — |
| `ecowitt_rain_rate_mm_per_hour` | — |
| `ecowitt_rain_mm` | period: event, hourly, daily, weekly, monthly, yearly, total |
| `ecowitt_solar_radiation_wm2` | — |
| `ecowitt_uv_index` | — |
| `ecowitt_battery` | sensor: ch1, ch2, ch3, wh65batt |
| `ecowitt_last_update_timestamp` | — |

---

## Tab 1 — Outdoor

### Canvas Panel (full width, ~200px tall)

Current conditions at a glance. All values are the latest point (`last_over_time` or direct gauge query).

**Layout:**
```
┌─────────────────────────────────────────────────────────────────────┐
│  🌡 Outdoor Temp    💧 Humidity    │  💨 Wind        ↗ Gust         │
│     XX.X °C           XX %         │     XX.X km/h    XX.X km/h      │
│                                    │     Direction: XXX°              │
├────────────────────────────────────┤                                  │
│  📊 Rel Pressure   Abs Pressure   │  ☀ Solar Rad    🔆 UV Index     │
│     XXXX.X hPa      XXXX.X hPa    │     XXX W/m²      X             │
│                                    │                                  │
│  🌧 Rain Rate: XX mm/h             │         Last update: HH:MM:SS   │
└─────────────────────────────────────────────────────────────────────┘
```

Elements: metric-value text boxes with labels, coloured backgrounds for weather context (blue tones for rain/humidity, amber for solar/UV, grey for pressure/wind).

### Time Series Panels (below canvas, 2 columns)

**Left column:**
1. Outdoor Temperature — `ecowitt_temperature_celsius{location="outdoor"}`
2. Outdoor Humidity — `ecowitt_humidity_percent{location="outdoor"}`
3. Pressure — both relative and absolute on same panel

**Right column:**
4. Wind Speed & Gust — speed and gust on same panel, max daily gust as separate line
5. Rain Accumulation — daily, weekly, monthly on same panel
6. Solar Radiation & UV — dual-axis: solar (W/m²) left, UV index right

---

## Tab 2 — Indoor

### Canvas Panel (full width, ~200px tall)

**Layout:**
```
┌─────────────────────────────────────────────────────────────────────┐
│  🏠 Indoor                        │  📡 Channel Sensors             │
│     Temp: XX.X °C                 │  Ch1: XX.X°C  XX%               │
│     Humidity: XX %                │  Ch2: XX.X°C  XX%               │
│                                   │  Ch3: XX.X°C  XX%               │
├───────────────────────────────────┴─────────────────────────────────┤
│  🔋 Batteries:  Ch1: X   Ch2: X   Ch3: X   WH65: X                 │
└─────────────────────────────────────────────────────────────────────┘
```

Elements: indoor temp/humidity prominently left, channel sensor grid right, battery row at bottom.

### Time Series Panels (below canvas, 2 columns)

**Left column:**
1. Indoor Temperature — `ecowitt_temperature_celsius{location="indoor"}`
2. Indoor Humidity — `ecowitt_humidity_percent{location="indoor"}`

**Right column:**
3. Channel Temperatures — ch1, ch2, ch3 on same panel
4. Channel Humidity — ch1, ch2, ch3 on same panel

---

## Dashboard Settings

| Setting | Value |
|---|---|
| Title | Ecowitt Weather Station |
| UID | `ecowitt-weather` |
| Refresh | 60s |
| Default time range | Last 24h |
| Datasource | Prometheus (UID: `prometheus`) |
| Tabs | Outdoor, Indoor |

---

## Delivery

- Dashboard JSON created via Grafana HTTP API (`POST /api/dashboards/db`)
- JSON also saved to `grafana/ecowitt-dashboard.json` in this repo for reproducibility
- Grafana at `http://localhost:3001`, credentials `admin/admin`
