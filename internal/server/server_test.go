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
)

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
	srv := New(config.Default(), testWebFS(t))
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
	cfg := config.Default()
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
	srv := New(config.Default(), testWebFS(t))

	paths := []string{"/css/main.css", "/js/app.js", "/js/tailwind.js"}

	for _, path := range paths {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		srv.Handler().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want %d", path, w.Code, http.StatusOK)
		}
	}
}
