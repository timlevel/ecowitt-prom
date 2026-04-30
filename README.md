# ecowitt-prom

A Prometheus exporter for Ecowitt weather stations (WS3900/WS3910 and compatible).

Receives weather data via the Ecowitt custom server protocol and exposes it as Prometheus metrics.

## How it works

1. Configure your Ecowitt station (via the WS View app) to POST data to this service
2. The service parses the form-encoded payload and updates Prometheus gauges
3. Prometheus scrapes `/metrics` to collect the data

All units are converted to metric (Celsius, km/h, hPa, mm) at exposure time.

## Quick Start

### Docker

```bash
docker compose up -d
```

### Binary

```bash
make build
./ecowitt-prom
```

## Configuration

| Environment Variable | Default | Description |
|---|---|---|
| `LISTEN_ADDR` | `:8080` | HTTP listen address |
| `DATA_PATH` | `/data/report` | Path the station POSTs to |
| `METRICS_PATH` | `/metrics` | Path Prometheus scrapes |
| `STATION_PASSKEY` | _(empty)_ | If set, only accept data from this PASSKEY |

## Ecowitt Station Configuration

In the WS View (or awnet) app:

- **Protocol:** Ecowitt
- **Server IP/Host:** Your server IP
- **Path:** `/data/report`
- **Port:** `8080`
- **Upload Interval:** 60 seconds (or your preference)

## Metrics

All metrics use the `ecowitt_` namespace with `station_type` and `model` labels.

| Metric | Description |
|---|---|
| `ecowitt_temperature_celsius` | Temperature (location: indoor/outdoor/ch1-ch8) |
| `ecowitt_humidity_percent` | Relative humidity |
| `ecowitt_pressure_hpa` | Barometric pressure (type: relative/absolute) |
| `ecowitt_wind_direction_degrees` | Wind direction |
| `ecowitt_wind_speed_kph` | Wind speed |
| `ecowitt_wind_gust_kph` | Wind gust speed |
| `ecowitt_max_daily_gust_kph` | Maximum daily gust |
| `ecowitt_rain_rate_mm_per_hour` | Rain rate |
| `ecowitt_rain_mm` | Rainfall accumulation (period: event/hourly/daily/weekly/monthly/yearly/total) |
| `ecowitt_solar_radiation_wm2` | Solar radiation |
| `ecowitt_uv_index` | UV index |
| `ecowitt_battery` | Sensor battery level |
| `ecowitt_soil_moisture_percent` | Soil moisture |
| `ecowitt_last_update_timestamp` | Unix timestamp of last data update |

## Endpoints

| Path | Method | Description |
|---|---|---|
| `/data/report` | POST | Receives Ecowitt station data |
| `/metrics` | GET | Prometheus metrics endpoint |
| `/healthz` | GET | Health check |

## License

MIT