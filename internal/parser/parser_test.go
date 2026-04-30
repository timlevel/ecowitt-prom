package parser

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestParseBasicPayload(t *testing.T) {
	form := url.Values{}
	form.Set("PASSKEY", "abc123def456")
	form.Set("stationtype", "EasyWeatherV1.5.9")
	form.Set("dateutc", "2021-08-17 00:15:59")
	form.Set("tempinf", "72.9")
	form.Set("humidityin", "62")
	form.Set("baromrelin", "29.829")
	form.Set("baromabsin", "28.122")
	form.Set("tempf", "22.0")
	form.Set("humidity", "100")
	form.Set("winddir", "271")
	form.Set("windspeedmph", "6.9")
	form.Set("windgustmph", "9.2")
	form.Set("maxdailygust", "9.2")
	form.Set("rainratein", "0.000")
	form.Set("eventrainin", "1.331")
	form.Set("hourlyrainin", "0.000")
	form.Set("dailyrainin", "0.000")
	form.Set("weeklyrainin", "1.331")
	form.Set("monthlyrainin", "4.929")
	form.Set("totalrainin", "14.890")
	form.Set("solarradiation", "0.00")
	form.Set("uv", "0")
	form.Set("wh65batt", "0")
	form.Set("freq", "868M")
	form.Set("model", "WS2900_V2.01.13")

	req := &http.Request{
		Method: http.MethodPost,
		Header: http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}},
		Body:   io.NopCloser(strings.NewReader(form.Encode())),
	}
	req = req.WithContext(req.Context())

	data, err := Parse(req)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if data.Passkey != "abc123def456" {
		t.Errorf("Passkey = %q, want %q", data.Passkey, "abc123def456")
	}
	if data.StationType != "EasyWeatherV1.5.9" {
		t.Errorf("StationType = %q, want %q", data.StationType, "EasyWeatherV1.5.9")
	}
	if data.Model != "WS2900_V2.01.13" {
		t.Errorf("Model = %q, want %q", data.Model, "WS2900_V2.01.13")
	}
	if data.TempIndoorF != 72.9 {
		t.Errorf("TempIndoorF = %v, want 72.9", data.TempIndoorF)
	}
	if data.HumidityIndoor != 62 {
		t.Errorf("HumidityIndoor = %v, want 62", data.HumidityIndoor)
	}
	if data.TempOutdoorF != 22.0 {
		t.Errorf("TempOutdoorF = %v, want 22.0", data.TempOutdoorF)
	}
	if data.HumidityOutdoor != 100 {
		t.Errorf("HumidityOutdoor = %v, want 100", data.HumidityOutdoor)
	}
	if data.WindDir != 271 {
		t.Errorf("WindDir = %v, want 271", data.WindDir)
	}
	if data.WindSpeedMPH != 6.9 {
		t.Errorf("WindSpeedMPH = %v, want 6.9", data.WindSpeedMPH)
	}
	if data.WindGustMPH != 9.2 {
		t.Errorf("WindGustMPH = %v, want 9.2", data.WindGustMPH)
	}
	if data.SolarRadiation != 0 {
		t.Errorf("SolarRadiation = %v, want 0", data.SolarRadiation)
	}
	if data.UV != 0 {
		t.Errorf("UV = %v, want 0", data.UV)
	}
	if data.EventRainIn != 1.331 {
		t.Errorf("EventRainIn = %v, want 1.331", data.EventRainIn)
	}
}

func TestParseChannelSensors(t *testing.T) {
	form := url.Values{}
	form.Set("PASSKEY", "abc123")
	form.Set("stationtype", "GW2000A_V2.1.4")
	form.Set("dateutc", "2022-04-20 19:14:47")
	form.Set("temp1f", "71.2")
	form.Set("humidity1", "61")
	form.Set("temp2f", "71.2")
	form.Set("humidity2", "58")
	form.Set("soilmoisture1", "53")
	form.Set("soilmoisture2", "57")
	form.Set("soilbatt1", "1.4")
	form.Set("soilbatt2", "1.3")

	req := &http.Request{
		Method: http.MethodPost,
		Header: http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}},
		Body:   io.NopCloser(strings.NewReader(form.Encode())),
	}
	req = req.WithContext(req.Context())

	data, err := Parse(req)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if data.ChannelTempsF[1] != 71.2 {
		t.Errorf("ChannelTempsF[1] = %v, want 71.2", data.ChannelTempsF[1])
	}
	if data.ChannelHumidity[1] != 61 {
		t.Errorf("ChannelHumidity[1] = %v, want 61", data.ChannelHumidity[1])
	}
	if data.ChannelTempsF[2] != 71.2 {
		t.Errorf("ChannelTempsF[2] = %v, want 71.2", data.ChannelTempsF[2])
	}
	if data.SoilMoisture[1] != 53 {
		t.Errorf("SoilMoisture[1] = %v, want 53", data.SoilMoisture[1])
	}
	if data.SoilBatteries[1] != 1.4 {
		t.Errorf("SoilBatteries[1] = %v, want 1.4", data.SoilBatteries[1])
	}
}

func TestUnitConversions(t *testing.T) {
	if got := FToC(32); got != 0 {
		t.Errorf("FToC(32) = %v, want 0", got)
	}
	if got := FToC(212); got != 100 {
		t.Errorf("FToC(212) = %v, want 100", got)
	}
	if got := FToC(72.9); got < 22.6 || got > 22.8 {
		t.Errorf("FToC(72.9) = %v, want ~22.72", got)
	}

	if got := InHgToHPA(29.9213); got < 1013.2 || got > 1013.3 {
		t.Errorf("InHgToHPA(29.9213) = %v, want ~1013.25", got)
	}

	if got := MPHToKPH(62.1371); got < 99.9 || got > 100.1 {
		t.Errorf("MPHToKPH(62.1371) = %v, want ~100", got)
	}

	if got := InchesToMM(1.0); got != 25.4 {
		t.Errorf("InchesToMM(1.0) = %v, want 25.4", got)
	}
}

func TestHasField(t *testing.T) {
	data := &EcowittData{
		FieldsPresent: map[string]string{"tempf": "72.9", "humidity": "65"},
	}
	if !HasField(data, "tempf") {
		t.Error("HasField(tempf) = false, want true")
	}
	if HasField(data, "winddir") {
		t.Error("HasField(winddir) = true, want false")
	}
}
