package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/timlevett/ecowitt-prom/internal/parser"
)

type Exporter struct {
	temperatureGauge    *prometheus.GaugeVec
	humidityGauge       *prometheus.GaugeVec
	pressureGauge       *prometheus.GaugeVec
	windDirGauge        *prometheus.GaugeVec
	windSpeedGauge      *prometheus.GaugeVec
	windGustGauge       *prometheus.GaugeVec
	maxDailyGustGauge   *prometheus.GaugeVec
	rainRateGauge       *prometheus.GaugeVec
	rainGauge           *prometheus.GaugeVec
	solarRadiationGauge *prometheus.GaugeVec
	uvGauge             *prometheus.GaugeVec
	batteryGauge        *prometheus.GaugeVec
	soilMoistureGauge   *prometheus.GaugeVec
	lastUpdateGauge     *prometheus.GaugeVec
}

func NewExporter(reg prometheus.Registerer) *Exporter {
	e := &Exporter{
		temperatureGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "ecowitt",
			Name:      "temperature_celsius",
			Help:      "Temperature in Celsius",
		}, []string{"station_type", "model", "location"}),
		humidityGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "ecowitt",
			Name:      "humidity_percent",
			Help:      "Relative humidity percentage",
		}, []string{"station_type", "model", "location"}),
		pressureGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "ecowitt",
			Name:      "pressure_hpa",
			Help:      "Barometric pressure in hPa",
		}, []string{"station_type", "model", "type"}),
		windDirGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "ecowitt",
			Name:      "wind_direction_degrees",
			Help:      "Wind direction in degrees",
		}, []string{"station_type", "model"}),
		windSpeedGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "ecowitt",
			Name:      "wind_speed_kph",
			Help:      "Wind speed in km/h",
		}, []string{"station_type", "model"}),
		windGustGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "ecowitt",
			Name:      "wind_gust_kph",
			Help:      "Wind gust speed in km/h",
		}, []string{"station_type", "model"}),
		maxDailyGustGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "ecowitt",
			Name:      "max_daily_gust_kph",
			Help:      "Maximum daily gust speed in km/h",
		}, []string{"station_type", "model"}),
		rainRateGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "ecowitt",
			Name:      "rain_rate_mm_per_hour",
			Help:      "Rain rate in mm/h",
		}, []string{"station_type", "model"}),
		rainGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "ecowitt",
			Name:      "rain_mm",
			Help:      "Rainfall accumulation in mm",
		}, []string{"station_type", "model", "period"}),
		solarRadiationGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "ecowitt",
			Name:      "solar_radiation_wm2",
			Help:      "Solar radiation in W/m^2",
		}, []string{"station_type", "model"}),
		uvGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "ecowitt",
			Name:      "uv_index",
			Help:      "UV index",
		}, []string{"station_type", "model"}),
		batteryGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "ecowitt",
			Name:      "battery",
			Help:      "Sensor battery level",
		}, []string{"station_type", "model", "sensor"}),
		soilMoistureGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "ecowitt",
			Name:      "soil_moisture_percent",
			Help:      "Soil moisture percentage",
		}, []string{"station_type", "model", "channel"}),
		lastUpdateGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "ecowitt",
			Name:      "last_update_timestamp",
			Help:      "Unix timestamp of last data update from station",
		}, []string{"station_type", "model"}),
	}

	if reg != nil {
		reg.MustRegister(
			e.temperatureGauge,
			e.humidityGauge,
			e.pressureGauge,
			e.windDirGauge,
			e.windSpeedGauge,
			e.windGustGauge,
			e.maxDailyGustGauge,
			e.rainRateGauge,
			e.rainGauge,
			e.solarRadiationGauge,
			e.uvGauge,
			e.batteryGauge,
			e.soilMoistureGauge,
			e.lastUpdateGauge,
		)
	}

	return e
}

func (e *Exporter) Update(data *parser.EcowittData) {
	stype := data.StationType
	model := data.Model

	e.deleteStaleLabels(e.temperatureGauge, []string{"station_type", "model", "location"}, stype, model)
	e.deleteStaleLabels(e.humidityGauge, []string{"station_type", "model", "location"}, stype, model)
	e.deleteStaleLabels(e.pressureGauge, []string{"station_type", "model", "type"}, stype, model)
	e.deleteStaleLabels(e.rainGauge, []string{"station_type", "model", "period"}, stype, model)
	e.deleteStaleLabels(e.batteryGauge, []string{"station_type", "model", "sensor"}, stype, model)
	e.deleteStaleLabels(e.soilMoistureGauge, []string{"station_type", "model", "channel"}, stype, model)

	if parser.HasField(data, "tempinf") {
		e.temperatureGauge.WithLabelValues(stype, model, "indoor").Set(parser.FToC(data.TempIndoorF))
	}
	if parser.HasField(data, "tempf") {
		e.temperatureGauge.WithLabelValues(stype, model, "outdoor").Set(parser.FToC(data.TempOutdoorF))
	}
	for ch, temp := range data.ChannelTempsF {
		e.temperatureGauge.WithLabelValues(stype, model, "ch"+string(rune('0'+ch))).Set(parser.FToC(temp))
	}

	if parser.HasField(data, "humidityin") {
		e.humidityGauge.WithLabelValues(stype, model, "indoor").Set(data.HumidityIndoor)
	}
	if parser.HasField(data, "humidity") {
		e.humidityGauge.WithLabelValues(stype, model, "outdoor").Set(data.HumidityOutdoor)
	}
	for ch, hum := range data.ChannelHumidity {
		e.humidityGauge.WithLabelValues(stype, model, "ch"+string(rune('0'+ch))).Set(hum)
	}

	if parser.HasField(data, "baromrelin") {
		e.pressureGauge.WithLabelValues(stype, model, "relative").Set(parser.InHgToHPA(data.BaromRelIn))
	}
	if parser.HasField(data, "baromabsin") {
		e.pressureGauge.WithLabelValues(stype, model, "absolute").Set(parser.InHgToHPA(data.BaromAbsIn))
	}

	if parser.HasField(data, "winddir") {
		e.windDirGauge.WithLabelValues(stype, model).Set(data.WindDir)
	}
	if parser.HasField(data, "windspeedmph") {
		e.windSpeedGauge.WithLabelValues(stype, model).Set(parser.MPHToKPH(data.WindSpeedMPH))
	}
	if parser.HasField(data, "windgustmph") {
		e.windGustGauge.WithLabelValues(stype, model).Set(parser.MPHToKPH(data.WindGustMPH))
	}
	if parser.HasField(data, "maxdailygust") {
		e.maxDailyGustGauge.WithLabelValues(stype, model).Set(parser.MPHToKPH(data.MaxDailyGustMPH))
	}

	if parser.HasField(data, "rainratein") {
		e.rainRateGauge.WithLabelValues(stype, model).Set(parser.InchesToMM(data.RainRateIn))
	}

	rainPeriods := []struct {
		field  string
		value  float64
		period string
	}{
		{"eventrainin", data.EventRainIn, "event"},
		{"hourlyrainin", data.HourlyRainIn, "hourly"},
		{"dailyrainin", data.DailyRainIn, "daily"},
		{"weeklyrainin", data.WeeklyRainIn, "weekly"},
		{"monthlyrainin", data.MonthlyRainIn, "monthly"},
		{"yearlyrainin", data.YearlyRainIn, "yearly"},
		{"totalrainin", data.TotalRainIn, "total"},
	}
	for _, rp := range rainPeriods {
		if parser.HasField(data, rp.field) {
			e.rainGauge.WithLabelValues(stype, model, rp.period).Set(parser.InchesToMM(rp.value))
		}
	}

	if parser.HasField(data, "solarradiation") {
		e.solarRadiationGauge.WithLabelValues(stype, model).Set(data.SolarRadiation)
	}
	if parser.HasField(data, "uv") {
		e.uvGauge.WithLabelValues(stype, model).Set(data.UV)
	}

	for sensor, val := range data.SensorBatteries {
		e.batteryGauge.WithLabelValues(stype, model, sensor).Set(val)
	}
	for ch, val := range data.ChannelBatteries {
		e.batteryGauge.WithLabelValues(stype, model, "ch"+string(rune('0'+ch))).Set(val)
	}

	for ch, val := range data.SoilMoisture {
		e.soilMoistureGauge.WithLabelValues(stype, model, "ch"+string(rune('0'+ch))).Set(val)
	}

	e.lastUpdateGauge.WithLabelValues(stype, model).SetToCurrentTime()
}

func (e *Exporter) deleteStaleLabels(gauge *prometheus.GaugeVec, labelNames []string, stype, model string) {
	var baseLabels prometheus.Labels
	switch len(labelNames) {
	case 3:
		key := labelNames[2]
		switch key {
		case "location":
			baseLabels = prometheus.Labels{"station_type": stype, "model": model, "location": ""}
		case "type":
			baseLabels = prometheus.Labels{"station_type": stype, "model": model, "type": ""}
		case "period":
			baseLabels = prometheus.Labels{"station_type": stype, "model": model, "period": ""}
		case "sensor":
			baseLabels = prometheus.Labels{"station_type": stype, "model": model, "sensor": ""}
		case "channel":
			baseLabels = prometheus.Labels{"station_type": stype, "model": model, "channel": ""}
		}
	default:
		return
	}
	if baseLabels != nil {
		gauge.Delete(baseLabels)
	}
}
