# ecowitt-prom Architecture Plan

## Overview

ecowitt-prom is a lightweight HTTP service that receives weather data from an Ecowitt WS3900/WS3910 weather station via the custom server protocol and exposes it as Prometheus metrics.

## Ecowitt Custom Server Protocol Summary

The Ecowitt gateway sends `POST` requests with `Content-Type: application/x-www-form-urlencoded` to a configurable endpoint. There is **no authentication** — the `PASSKEY` field is just a MAC address identifier, not a security token.

Typical WS3900 payload:

```
PASSKEY=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx&stationtype=EasyWeatherV1.5.9&dateutc=2021-08-17+00%3A15%3A59&tempinf=72.9&humidityin=62&baromrelin=29.829&baromabsin=28.122&tempf=22.0&humidity=100&winddir=271&windspeedmph=6.9&windgustmph=9.2&maxdailygust=9.2&rainratein=0.000&eventrainin=1.331&hourlyrainin=0.000&dailyrainin=0.000&weeklyrainin=1.331&monthlyrainin=4.929&totalrainin=14.890&solarradiation=0.00&uv=0&wh65batt=0&freq=868M&model=WS2900_V2.01.13
```

The station uploads at a configurable interval (default 60s, minimum ~16s).

## Architecture

### Stack

- **Language:** Go (single binary, no runtime dependencies, ideal for deployment on Raspberry Pi / home servers)
- **HTTP Framework:** `net/http` stdlib (minimal dependencies)
- **Metrics:** Prometheus client_golang for exposition
- **Configuration:** Environment variables (12-factor style) with sensible defaults

### Components

```
┌──────────────┐       POST /data/report        ┌──────────────────┐       GET /metrics       ┌─────────────┐
│  Ecowitt GW  │ ──────────────────────────────► │  ecowitt-prom    │ ────────────────────────► │  Prometheus  │
│  (station)   │   application/x-www-form-        │  (this service)  │   text/prometheus        │  (server)    │
│              │   urlencoded                     │                  │                          │              │
└──────────────┘                                  └──────────────────┘                          └─────────────┘
```

### Data Flow

1. Ecowitt gateway POSTs form-encoded data to `/data/report`
2. Parser extracts metric values from form fields
3. Values are stored in Prometheus gauge/counter objects
4. Prometheus scrapes `/metrics` on its interval

### Configuration

| Env Var | Default | Description |
|---|---|---|
| `LISTEN_ADDR` | `:8080` | HTTP listen address |
| `DATA_PATH` | `/data/report` | Path the station POSTs to |
| `METRICS_PATH` | `/metrics` | Path Prometheus scrapes |
| `STATION_PASSKEY` | _(empty)_ | If set, only accept data from this PASSKEY (light validation) |

### Metric Design

All metrics use the `ecowitt_` namespace. Labels include `station_type` and `model` from the payload. Unit conversions to metric (Celsius, km/h, hPa, mm) happen at exposure time — the station sends imperial by default.

**Core weather metrics:**

| Prometheus Metric | Type | Ecowitt Field(s) | Notes |
|---|---|---|---|
| `ecowitt_temperature_celsius` | Gauge | `tempf`, `tempinf`, `temp1f`..`temp8f`, `tf_co2` | Converted from F to C |
| `ecowitt_humidity_percent` | Gauge | `humidity`, `humidityin`, `humidity1`..`humidity8`, `humi_co2` | |
| `ecowitt_pressure_hpa` | Gauge | `baromrelin`, `baromabsin` | Converted from inHg to hPa |
| `ecowitt_wind_direction_degrees` | Gauge | `winddir` | |
| `ecowitt_wind_speed_kph` | Gauge | `windspeedmph` | Converted from mph to km/h |
| `ecowitt_wind_gust_kph` | Gauge | `windgustmph` | Converted from mph to km/h |
| `ecowitt_max_daily_gust_kph` | Gauge | `maxdailygust` | |
| `ecowitt_rain_rate_mm_per_hour` | Gauge | `rainratein` | Converted from in/h to mm/h |
| `ecowitt_rain_mm` | Gauge | `eventrainin`, `hourlyrainin`, `dailyrainin`, `weeklyrainin`, `monthlyrainin`, `yearlyrainin`, `totalrainin` | `period` label |
| `ecowitt_solar_radiation_wm2` | Gauge | `solarradiation` | |
| `ecowitt_uv_index` | Gauge | `uv` | |
| `ecowitt_battery` | Gauge | `wh65batt`, `wh25batt`, etc. | `sensor` label |

Labels on metrics:
- `station_type` — from `stationtype` field
- `model` — from `model` field
- `location` — `outdoor`/`indoor`/`ch1`..`ch8` for multi-channel sensors
- `period` — for rain gauges: `event`/`hourly`/`daily`/`weekly`/`monthly`/`yearly`/`total`

### Project Structure

```
ecowitt-prom/
├── main.go                  # Entry point, config, HTTP server setup
├── internal/
│   ├── parser/
│   │   └── parser.go        # Parse form-encoded Ecowitt payloads
│   ├── metrics/
│   │   └── metrics.go        # Prometheus metric definitions and updates
│   └── server/
│       └── server.go         # HTTP handlers, routing
├── go.mod
├── go.sum
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── README.md
└── .gitignore
```

### Key Design Decisions

1. **Metric-first conversion**: Convert imperial→metric at exposure time so Prometheus always scrapes metric units. This is the common convention for Prometheus exporters.

2. **Ignore unknown fields gracefully**: New Ecowitt sensors may add fields. The parser should skip fields it doesn't recognize rather than erroring.

3. **No persistence**: This is a stateless exporter. If it restarts, Prometheus will have a gap. The Ecowitt station re-sends data on its interval so the exporter self-heals.

4. **Single station design**: Initial version supports one station. Multi-station support can be added later via `passkey` label if needed.

5. **Gauges for everything**: Rain counters (hourly/daily/etc.) come from the station's own accumulation logic. We expose them as gauges rather than Prometheus counters because the station manages the lifecycle.

6. **Health check**: Expose `/healthz` for liveness probes.

## Deployment

### Docker

```yaml
# docker-compose.yml
services:
  ecowitt-prom:
    build: .
    ports:
      - "8080:8080"
    environment:
      - LISTEN_ADDR=:8080
    restart: unless-stopped
```

### Manual

```bash
go build -o ecowitt-prom .
./ecowitt-prom
```

### Ecowitt Station Configuration

In the WS View app, configure:
- Protocol: Ecowitt
- Server IP/Host: `<your-server-ip>`
- Path: `/data/report`
- Port: `8080`
- Upload Interval: `60` (seconds)

## Future Considerations (out of scope for Phase 1)

- Multi-station support (passkey label)
- HTTPS/TLS termination (use reverse proxy)
- Webhook/forwarding to Ecowitt.net simultaneously
- Metric unit configuration (imperial vs metric output)
- Piezo rain sensor fields (WS3910)
- Lightning, air quality, soil moisture, leak sensor support