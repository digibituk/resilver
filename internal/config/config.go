package config

import (
	"encoding/json"
	"errors"
	"os"
)

type Config struct {
	Server  ServerConfig            `json:"server"`
	Layout  LayoutConfig            `json:"layout"`
	Modules map[string]ModuleConfig `json:"modules"`
}

type ServerConfig struct {
	Port int `json:"port"`
}

type LayoutConfig struct {
	Columns   int                 `json:"columns"`
	Rows      int                 `json:"rows"`
	Positions map[string][]string `json:"positions"`
}

type ModuleConfig struct {
	Enabled bool           `json:"enabled"`
	Config  map[string]any `json:"config"`
}

func Default() Config {
	return Config{
		Server: ServerConfig{Port: 8080},
		Layout: LayoutConfig{
			Columns: 3,
			Rows:    3,
			Positions: map[string][]string{
				"top-left":      {},
				"top-center":    {"clock"},
				"top-right":     {"weather"},
				"middle-left":   {},
				"middle-center": {},
				"middle-right":  {},
				"bottom-left":   {},
				"bottom-center": {},
				"bottom-right":  {},
			},
		},
		Modules: map[string]ModuleConfig{
			"clock": {
				Enabled: true,
				Config: map[string]any{
					"format":      "24h",
					"showSeconds": true,
					"showDate":    true,
				},
			},
			"weather": {
				Enabled: true,
				Config: map[string]any{
					"latitude":               51.4778356052696,
					"longitude":              0.323272352543598,
					"units":                  "celsius",
					"location":               "Grays",
					"refreshIntervalSeconds": 1800,
				},
			},
		},
	}
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Default(), nil
		}
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
