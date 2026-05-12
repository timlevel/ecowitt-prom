package parser

import (
	"net/http"
	"strconv"
	"strings"
)

type EcowittData struct {
	Passkey        string
	StationType    string
	Model          string
	DateUTC        string
	Freq           string
	UploadInterval int

	TempIndoorF    float64
	HumidityIndoor float64
	BaromRelIn     float64
	BaromAbsIn     float64

	TempOutdoorF    float64
	HumidityOutdoor float64
	WindDir         float64
	WindSpeedMPH    float64
	WindGustMPH     float64
	MaxDailyGustMPH float64
	SolarRadiation  float64
	UV              float64
	VPDkPa          float64

	RainRateIn    float64
	EventRainIn   float64
	HourlyRainIn  float64
	DailyRainIn   float64
	WeeklyRainIn  float64
	MonthlyRainIn float64
	YearlyRainIn  float64
	TotalRainIn   float64

	ChannelTempsF    map[int]float64
	ChannelHumidity  map[int]float64
	ChannelBatteries map[int]float64

	SoilMoisture  map[int]float64
	SoilBatteries map[int]float64

	SensorBatteries map[string]float64

	FieldsPresent map[string]string
}

func Parse(r *http.Request) (*EcowittData, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	data := &EcowittData{
		ChannelTempsF:    make(map[int]float64),
		ChannelHumidity:  make(map[int]float64),
		ChannelBatteries: make(map[int]float64),
		SoilMoisture:     make(map[int]float64),
		SoilBatteries:    make(map[int]float64),
		SensorBatteries:  make(map[string]float64),
		FieldsPresent:    make(map[string]string),
	}

	for key, values := range r.Form {
		if len(values) > 0 {
			data.FieldsPresent[key] = values[0]
		}
	}

	data.Passkey = data.FieldsPresent["PASSKEY"]
	data.StationType = data.FieldsPresent["stationtype"]
	data.Model = data.FieldsPresent["model"]
	data.DateUTC = data.FieldsPresent["dateutc"]
	data.Freq = data.FieldsPresent["freq"]

	if v, ok := parseFloat(data.FieldsPresent, "interval"); ok {
		data.UploadInterval = int(v)
	}

	data.TempIndoorF, _ = parseFloat(data.FieldsPresent, "tempinf")
	data.HumidityIndoor, _ = parseFloat(data.FieldsPresent, "humidityin")
	data.BaromRelIn, _ = parseFloat(data.FieldsPresent, "baromrelin")
	data.BaromAbsIn, _ = parseFloat(data.FieldsPresent, "baromabsin")

	data.TempOutdoorF, _ = parseFloat(data.FieldsPresent, "tempf")
	data.HumidityOutdoor, _ = parseFloat(data.FieldsPresent, "humidity")
	data.WindDir, _ = parseFloat(data.FieldsPresent, "winddir")
	data.WindSpeedMPH, _ = parseFloat(data.FieldsPresent, "windspeedmph")
	data.WindGustMPH, _ = parseFloat(data.FieldsPresent, "windgustmph")
	data.MaxDailyGustMPH, _ = parseFloat(data.FieldsPresent, "maxdailygust")
	data.SolarRadiation, _ = parseFloat(data.FieldsPresent, "solarradiation")
	data.UV, _ = parseFloat(data.FieldsPresent, "uv")
	data.VPDkPa, _ = parseFloat(data.FieldsPresent, "vpd")

	data.RainRateIn, _ = parseFloat(data.FieldsPresent, "rainratein")
	data.EventRainIn, _ = parseFloat(data.FieldsPresent, "eventrainin")
	data.HourlyRainIn, _ = parseFloat(data.FieldsPresent, "hourlyrainin")
	data.DailyRainIn, _ = parseFloat(data.FieldsPresent, "dailyrainin")
	data.WeeklyRainIn, _ = parseFloat(data.FieldsPresent, "weeklyrainin")
	data.MonthlyRainIn, _ = parseFloat(data.FieldsPresent, "monthlyrainin")
	data.YearlyRainIn, _ = parseFloat(data.FieldsPresent, "yearlyrainin")
	data.TotalRainIn, _ = parseFloat(data.FieldsPresent, "totalrainin")

	for i := 1; i <= 8; i++ {
		tempKey := "temp" + strconv.Itoa(i) + "f"
		humKey := "humidity" + strconv.Itoa(i)
		battKey := "batt" + strconv.Itoa(i)

		if v, ok := parseFloat(data.FieldsPresent, tempKey); ok {
			data.ChannelTempsF[i] = v
		}
		if v, ok := parseFloat(data.FieldsPresent, humKey); ok {
			data.ChannelHumidity[i] = v
		}
		if v, ok := parseFloat(data.FieldsPresent, battKey); ok {
			data.ChannelBatteries[i] = v
		}
	}

	for i := 1; i <= 8; i++ {
		smKey := "soilmoisture" + strconv.Itoa(i)
		sbKey := "soilbatt" + strconv.Itoa(i)

		if v, ok := parseFloat(data.FieldsPresent, smKey); ok {
			data.SoilMoisture[i] = v
		}
		if v, ok := parseFloat(data.FieldsPresent, sbKey); ok {
			data.SoilBatteries[i] = v
		}
	}

	batterySensors := []string{
		"wh25batt", "wh26batt", "wh40batt", "wh57batt",
		"wh65batt", "wh68batt", "wh80batt", "wh90batt",
		"wh90battpc", "pm25batt1", "co2_batt",
	}

	for _, sensor := range batterySensors {
		if v, ok := parseFloat(data.FieldsPresent, sensor); ok {
			data.SensorBatteries[sensor] = v
		}
	}

	return data, nil
}

func FToC(f float64) float64 {
	return (f - 32) * 5 / 9
}

func InHgToHPA(inHg float64) float64 {
	return inHg * 33.8639
}

func MPHToKPH(mph float64) float64 {
	return mph * 1.60934
}

func InchesToMM(inches float64) float64 {
	return inches * 25.4
}

func HasField(data *EcowittData, field string) bool {
	_, ok := data.FieldsPresent[field]
	return ok
}

func parseFloat(fields map[string]string, key string) (float64, bool) {
	v, ok := fields[key]
	if !ok || v == "" {
		return 0, false
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

func ParseBatteryKey(key string) (string, bool) {
	batteryPrefixes := []string{"wh25batt", "wh26batt", "wh40batt", "wh57batt", "wh65batt", "wh68batt", "wh80batt", "wh90batt", "wh90battpc", "pm25batt1", "co2_batt"}

	for _, prefix := range batteryPrefixes {
		if strings.HasPrefix(key, prefix) {
			return prefix, true
		}
	}
	return "", false
}
