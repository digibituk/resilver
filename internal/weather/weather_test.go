package weather

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func fakeOpenMeteoServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := `{
			"current": {
				"temperature_2m": 15.3,
				"relative_humidity_2m": 72,
				"apparent_temperature": 13.1,
				"weather_code": 3,
				"wind_speed_10m": 12.5
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(resp))
	}))
}

func TestFetchCurrentWeather(t *testing.T) {
	srv := fakeOpenMeteoServer(t)
	defer srv.Close()

	client := NewClient(srv.URL)
	data, err := client.Fetch(51.5074, -0.1278, "celsius")
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	if data.Temperature != 15.3 {
		t.Errorf("Temperature = %f, want 15.3", data.Temperature)
	}
	if data.Humidity != 72 {
		t.Errorf("Humidity = %d, want 72", data.Humidity)
	}
	if data.ApparentTemp != 13.1 {
		t.Errorf("ApparentTemp = %f, want 13.1", data.ApparentTemp)
	}
	if data.WeatherCode != 3 {
		t.Errorf("WeatherCode = %d, want 3", data.WeatherCode)
	}
	if data.WindSpeed != 12.5 {
		t.Errorf("WindSpeed = %f, want 12.5", data.WindSpeed)
	}
	if data.Description == "" {
		t.Error("Description should not be empty")
	}
	if data.Icon == "" {
		t.Error("Icon should not be empty")
	}
}

func TestFetchReturnsErrorOnBadResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	_, err := client.Fetch(51.5074, -0.1278, "celsius")
	if err == nil {
		t.Error("Fetch() expected error on 500, got nil")
	}
}

func TestFetchReturnsErrorOnInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{invalid`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	_, err := client.Fetch(51.5074, -0.1278, "celsius")
	if err == nil {
		t.Error("Fetch() expected error on invalid JSON, got nil")
	}
}

func TestWeatherCodeMapping(t *testing.T) {
	tests := []struct {
		code        int
		description string
		icon        string
	}{
		{0, "Clear sky", "☀️"},
		{1, "Mainly clear", "🌤️"},
		{45, "Foggy", "🌫️"},
		{61, "Light rain", "🌧️"},
		{95, "Thunderstorm", "⛈️"},
		{999, "Unknown", "❓"},
	}

	for _, tc := range tests {
		desc, icon := describeWeatherCode(tc.code)
		if desc != tc.description {
			t.Errorf("describeWeatherCode(%d) description = %q, want %q", tc.code, desc, tc.description)
		}
		if icon != tc.icon {
			t.Errorf("describeWeatherCode(%d) icon = %q, want %q", tc.code, icon, tc.icon)
		}
	}
}

func TestCachedClientReturnsCachedData(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := `{"current":{"temperature_2m":15.3,"relative_humidity_2m":72,"apparent_temperature":13.1,"weather_code":3,"wind_speed_10m":12.5}}`
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(resp))
	}))
	defer srv.Close()

	client := NewCachedClient(srv.URL, 1*time.Minute)

	// First call hits the server
	d1, err := client.Fetch(51.5074, -0.1278, "celsius")
	if err != nil {
		t.Fatalf("first Fetch() error: %v", err)
	}

	// Second call should use cache
	d2, err := client.Fetch(51.5074, -0.1278, "celsius")
	if err != nil {
		t.Fatalf("second Fetch() error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected 1 upstream call, got %d", callCount)
	}
	if d1.Temperature != d2.Temperature {
		t.Error("cached data should match original")
	}
}

func TestWeatherDataJSON(t *testing.T) {
	data := WeatherData{
		Temperature:  15.3,
		ApparentTemp: 13.1,
		Humidity:     72,
		WeatherCode:  3,
		WindSpeed:    12.5,
		Description:  "Overcast",
		Icon:         "☁️",
	}

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var got WeatherData
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if got.Temperature != 15.3 {
		t.Errorf("Temperature = %f, want 15.3", got.Temperature)
	}
}
