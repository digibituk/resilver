package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	data := []byte(`{
		"server": {"port": 9090},
		"layout": {
			"columns": 2,
			"rows": 2,
			"positions": {
				"top-left": ["clock"],
				"top-right": [],
				"bottom-left": [],
				"bottom-right": []
			}
		},
		"modules": {
			"clock": {
				"enabled": true,
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
	if cfg.Layout.Columns != 2 {
		t.Errorf("Layout.Columns = %d, want 2", cfg.Layout.Columns)
	}
	if cfg.Layout.Rows != 2 {
		t.Errorf("Layout.Rows = %d, want 2", cfg.Layout.Rows)
	}

	clock, ok := cfg.Modules["clock"]
	if !ok {
		t.Fatal("clock module not found")
	}
	if !clock.Enabled {
		t.Error("clock.Enabled = false, want true")
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
	if cfg.Layout.Columns != 3 {
		t.Errorf("Layout.Columns = %d, want default 3", cfg.Layout.Columns)
	}
	if cfg.Layout.Rows != 3 {
		t.Errorf("Layout.Rows = %d, want default 3", cfg.Layout.Rows)
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

func TestDefaultConfig(t *testing.T) {
	cfg := Default()

	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want 8080", cfg.Server.Port)
	}
	if cfg.Layout.Columns != 3 {
		t.Errorf("Layout.Columns = %d, want 3", cfg.Layout.Columns)
	}
	if len(cfg.Layout.Positions["top-center"]) != 1 || cfg.Layout.Positions["top-center"][0] != "clock" {
		t.Errorf("Layout.Positions[top-center] = %v, want [clock]", cfg.Layout.Positions["top-center"])
	}
}
