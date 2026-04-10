package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default()

	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want 8080", cfg.Server.Port)
	}
	if cfg.Layout.MaxWidgets != 8 {
		t.Errorf("Layout.MaxWidgets = %d, want 8", cfg.Layout.MaxWidgets)
	}
	if cfg.Layout.Direction != "row" {
		t.Errorf("Layout.Direction = %q, want row", cfg.Layout.Direction)
	}
	if len(cfg.Layout.Widgets) != 2 {
		t.Fatalf("Layout.Widgets length = %d, want 2", len(cfg.Layout.Widgets))
	}
	if cfg.Layout.Widgets[0].Module != "clock" {
		t.Errorf("Layout.Widgets[0].Module = %q, want clock", cfg.Layout.Widgets[0].Module)
	}
	if cfg.Layout.Widgets[1].Module != "weather" {
		t.Errorf("Layout.Widgets[1].Module = %q, want weather", cfg.Layout.Widgets[1].Module)
	}
}

func TestIsModuleActive(t *testing.T) {
	cfg := Default()

	if !cfg.IsModuleActive("clock") {
		t.Error("IsModuleActive(clock) = false, want true")
	}
	if !cfg.IsModuleActive("weather") {
		t.Error("IsModuleActive(weather) = false, want true")
	}
	if cfg.IsModuleActive("nonexistent") {
		t.Error("IsModuleActive(nonexistent) = true, want false")
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	data := []byte(`{
		"server": {"port": 9090},
		"layout": {
			"maxWidgets": 4,
			"direction": "column",
			"widgets": [
				{"module": "clock"}
			]
		},
		"modules": {
			"clock": {
				"config": {"format": "12h", "showSeconds": false, "showDate": true}
			}
		}
	}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("Server.Port = %d, want 9090", cfg.Server.Port)
	}
	if cfg.Layout.MaxWidgets != 4 {
		t.Errorf("Layout.MaxWidgets = %d, want 4", cfg.Layout.MaxWidgets)
	}
	if cfg.Layout.Direction != "column" {
		t.Errorf("Layout.Direction = %q, want column", cfg.Layout.Direction)
	}
	if len(cfg.Layout.Widgets) != 1 {
		t.Fatalf("Layout.Widgets length = %d, want 1", len(cfg.Layout.Widgets))
	}
	if cfg.Layout.Widgets[0].Module != "clock" {
		t.Errorf("Layout.Widgets[0].Module = %q, want clock", cfg.Layout.Widgets[0].Module)
	}

	clock, ok := cfg.Modules["clock"]
	if !ok {
		t.Fatal("clock module not found")
	}
	if clock.Config["format"] != "12h" {
		t.Errorf("clock format = %v, want 12h", clock.Config["format"])
	}
}

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.json")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want default 8080", cfg.Server.Port)
	}
	if cfg.Layout.MaxWidgets != 8 {
		t.Errorf("Layout.MaxWidgets = %d, want default 8", cfg.Layout.MaxWidgets)
	}
	if len(cfg.Layout.Widgets) != 2 {
		t.Errorf("Layout.Widgets length = %d, want default 2", len(cfg.Layout.Widgets))
	}
}

func TestLoadInvalidJSONReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	if err := os.WriteFile(path, []byte(`{invalid`), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("Load() expected error for invalid JSON, got nil")
	}
}

func TestLoadTimezoneField(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	data := []byte(`{
		"server": {"port": 8080},
		"layout": {
			"maxWidgets": 8,
			"direction": "row",
			"widgets": [{"module": "clock"}]
		},
		"modules": {
			"clock": {
				"config": {"format": "24h", "timezone": "Europe/London"}
			}
		}
	}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	clock := cfg.Modules["clock"]
	tz, ok := clock.Config["timezone"]
	if !ok {
		t.Fatal("timezone field not found in clock config")
	}
	if tz != "Europe/London" {
		t.Errorf("timezone = %v, want Europe/London", tz)
	}
}

func TestLoadEmptyModulesConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	data := []byte(`{
		"server": {"port": 8080},
		"layout": {
			"maxWidgets": 8,
			"direction": "row",
			"widgets": []
		},
		"modules": {}
	}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if len(cfg.Modules) != 0 {
		t.Errorf("Modules length = %d, want 0", len(cfg.Modules))
	}
	if len(cfg.Layout.Widgets) != 0 {
		t.Errorf("Layout.Widgets length = %d, want 0", len(cfg.Layout.Widgets))
	}
}

func TestLoadPermissionDeniedReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	if err := os.WriteFile(path, []byte(`{}`), 0000); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("Load() expected error for unreadable file, got nil")
	}
}

func TestLoadWidgetsExceedMaxReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	data := []byte(`{
		"server": {"port": 8080},
		"layout": {
			"maxWidgets": 2,
			"direction": "row",
			"widgets": [
				{"module": "a"},
				{"module": "b"},
				{"module": "c"}
			]
		},
		"modules": {}
	}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("Load() expected error when widgets exceed maxWidgets, got nil")
	}
}
