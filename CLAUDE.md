# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build      # compile binary (embeds version from git tag)
make test       # go test -v ./...
make run        # build + run
make docker     # docker build

go test -v ./internal/parser/...   # run parser tests only
```

## Architecture

Single-binary HTTP server. Data flow:

```
Ecowitt station POST → server.handleData → parser.Parse → metrics.Exporter.Update → Prometheus gauges
```

Three internal packages:

- **`internal/parser`** — parses `application/x-www-form-urlencoded` POSTs from the station. Raw values are imperial (°F, inHg, mph, inches); conversion functions (`FToC`, `InHgToHPA`, `MPHToKPH`, `InchesToMM`) live here and are called by the exporter. `FieldsPresent` map is the source of truth for which fields actually arrived — always check `HasField` before setting a gauge.
- **`internal/metrics`** — wraps `prometheus.GaugeVec` instances. Uses a custom (incomplete) `deleteStaleLabels` to reset label sets between updates. Metrics use an isolated `prometheus.Registry` (not the default global one), so default Go runtime metrics are not exposed.
- **`internal/server`** — functional-options pattern (`WithAddr`, `WithDataPath`, etc.). Wires the mux, optionally enforces `PASSKEY`, and owns the registry lifecycle.

## Key design constraints

- The station sends data in US customary units; all Prometheus metrics are metric units. Never expose imperial values.
- `metrics.deleteStaleLabels` currently deletes only a single label combination (not a full reset). Extending multi-value label sets (e.g. adding a new channel) requires understanding this limitation.
- Channel sensors (temp1f–temp8f, soilmoisture1–8) iterate 1–8 and use `string(rune('0'+ch))` for label suffix — produces "1"–"8" correctly only for single digits.

## Configuration

All config via env vars or flags (env takes precedence via `envOrDefault`):

| Env | Flag | Default |
|-----|------|---------|
| `LISTEN_ADDR` | `-addr` | `:8080` |
| `DATA_PATH` | `-data-path` | `/data/report` |
| `METRICS_PATH` | `-metrics-path` | `/metrics` |
| `STATION_PASSKEY` | `-passkey` | _(empty = accept all)_ |
