package server

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/digibituk/resilver/internal/config"
	"github.com/digibituk/resilver/internal/news"
	"github.com/digibituk/resilver/internal/weather"
)

const testRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>BBC News</title>
    <item><title>Breaking news headline</title><link>https://example.com/1</link></item>
    <item><title>Second story</title><link>https://example.com/2</link></item>
  </channel>
</rss>`

func fakeNewsServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(testRSS))
	}))
}

func fakeWeatherServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := `{"current":{"temperature_2m":15.3,"relative_humidity_2m":72,"apparent_temperature":13.1,"weather_code":3,"wind_speed_10m":12.5}}`
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(resp))
	}))
}

func loadTestConfig(t *testing.T) config.Config {
	t.Helper()
	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("failed to load default config: %v", err)
	}
	return cfg
}

func testWebFS(t *testing.T) fs.FS {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(file), "..", "..", "web")
	if _, err := os.Stat(root); err != nil {
		t.Fatalf("web directory not found at %s: %v", root, err)
	}
	return os.DirFS(root)
}

func TestServeIndex(t *testing.T) {
	srv := New(loadTestConfig(t), testWebFS(t))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET / status = %d, want %d", w.Code, http.StatusOK)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("GET / Content-Type = %q, want text/html", ct)
	}
	if !strings.Contains(w.Body.String(), "Resilver") {
		t.Error("GET / body does not contain 'Resilver'")
	}
}

func TestServeConfigEndpoint(t *testing.T) {
	cfg := loadTestConfig(t)
	srv := New(cfg, testWebFS(t))
	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /api/config status = %d, want %d", w.Code, http.StatusOK)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("GET /api/config Content-Type = %q, want application/json", ct)
	}

	var got config.Config
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode config response: %v", err)
	}
	if got.Server.Port != 8080 {
		t.Errorf("config.Server.Port = %d, want 8080", got.Server.Port)
	}
}

func TestServeStaticAssets(t *testing.T) {
	srv := New(loadTestConfig(t), testWebFS(t))

	tests := []struct {
		path        string
		contentType string
	}{
		{"/css/main.css", "text/css"},
		{"/js/app.js", "text/javascript"},
		{"/js/tailwind.js", "text/javascript"},
	}

	for _, tc := range tests {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		w := httptest.NewRecorder()
		srv.Handler().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want %d", tc.path, w.Code, http.StatusOK)
		}
		ct := w.Header().Get("Content-Type")
		if !strings.Contains(ct, tc.contentType) {
			t.Errorf("GET %s Content-Type = %q, want %s", tc.path, ct, tc.contentType)
		}
	}
}

func TestServe404ForMissingAsset(t *testing.T) {
	srv := New(loadTestConfig(t), testWebFS(t))
	req := httptest.NewRequest(http.MethodGet, "/nonexistent.js", nil)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GET /nonexistent.js status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestConfigEndpointReflectsCustomConfig(t *testing.T) {
	cfg := loadTestConfig(t)
	cfg.Server.Port = 9999

	srv := New(cfg, testWebFS(t))
	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	var got config.Config
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode config response: %v", err)
	}
	if got.Server.Port != 9999 {
		t.Errorf("config.Server.Port = %d, want 9999", got.Server.Port)
	}
}

func TestWeatherEndpointReturnsData(t *testing.T) {
	ws := fakeWeatherServer(t)
	defer ws.Close()

	cfg := loadTestConfig(t)

	srv := NewWithWeatherURL(cfg, testWebFS(t), ws.URL)
	req := httptest.NewRequest(http.MethodGet, "/api/weather", nil)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /api/weather status = %d, want %d", w.Code, http.StatusOK)
	}

	var data weather.WeatherData
	if err := json.Unmarshal(w.Body.Bytes(), &data); err != nil {
		t.Fatalf("failed to decode weather response: %v", err)
	}
	if data.Temperature != 15.3 {
		t.Errorf("Temperature = %f, want 15.3", data.Temperature)
	}
	if data.Description == "" {
		t.Error("Description should not be empty")
	}
}

func TestWeatherEndpoint404WhenNotInLayout(t *testing.T) {
	cfg := loadTestConfig(t)
	cfg.Layout.Widgets = []config.WidgetEntry{{Module: "clock"}}
	delete(cfg.Modules, "weather")

	srv := New(cfg, testWebFS(t))
	req := httptest.NewRequest(http.MethodGet, "/api/weather", nil)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GET /api/weather status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestWeatherEndpoint502OnUpstreamError(t *testing.T) {
	ws := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ws.Close()

	cfg := loadTestConfig(t)

	srv := NewWithWeatherURL(cfg, testWebFS(t), ws.URL)
	req := httptest.NewRequest(http.MethodGet, "/api/weather", nil)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Errorf("GET /api/weather status = %d, want %d", w.Code, http.StatusBadGateway)
	}
}

func TestNewsEndpointReturnsData(t *testing.T) {
	ns := fakeNewsServer(t)
	defer ns.Close()

	cfg := loadTestConfig(t)
	cfg.Layout.Widgets = append(cfg.Layout.Widgets, config.WidgetEntry{Module: "news"})
	cfg.Modules["news"] = config.ModuleConfig{
		Config: map[string]any{
			"feedUrls":               []any{ns.URL},
			"maxItems":               float64(5),
			"refreshIntervalSeconds": float64(1800),
		},
	}

	srv := New(cfg, testWebFS(t))
	req := httptest.NewRequest(http.MethodGet, "/api/news", nil)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /api/news status = %d, want %d", w.Code, http.StatusOK)
	}

	var items []news.NewsItem
	if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
		t.Fatalf("failed to decode news response: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].Title != "Breaking news headline" {
		t.Errorf("items[0].Title = %q, want Breaking news headline", items[0].Title)
	}
	if items[0].Source != "BBC News" {
		t.Errorf("items[0].Source = %q, want BBC News", items[0].Source)
	}
}

func TestNewsEndpoint404WhenNotInLayout(t *testing.T) {
	cfg := loadTestConfig(t)
	cfg.Layout.Widgets = []config.WidgetEntry{{Module: "clock"}, {Module: "weather"}}
	delete(cfg.Modules, "news")

	srv := New(cfg, testWebFS(t))
	req := httptest.NewRequest(http.MethodGet, "/api/news", nil)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GET /api/news status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestNewsEndpoint502OnUpstreamError(t *testing.T) {
	ns := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ns.Close()

	cfg := loadTestConfig(t)
	cfg.Layout.Widgets = append(cfg.Layout.Widgets, config.WidgetEntry{Module: "news"})
	cfg.Modules["news"] = config.ModuleConfig{
		Config: map[string]any{
			"feedUrls":               []any{ns.URL},
			"maxItems":               float64(5),
			"refreshIntervalSeconds": float64(1800),
		},
	}

	srv := New(cfg, testWebFS(t))
	req := httptest.NewRequest(http.MethodGet, "/api/news", nil)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Errorf("GET /api/news status = %d, want %d", w.Code, http.StatusBadGateway)
	}
}
