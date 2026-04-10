package config

import (
	"encoding/json"
	"errors"
	"fmt"
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
	MaxWidgets int           `json:"maxWidgets"`
	Direction  string        `json:"direction"`
	Widgets    []WidgetEntry `json:"widgets"`
}

type WidgetEntry struct {
	Module string `json:"module"`
}

type ModuleConfig struct {
	Config map[string]any `json:"config"`
}

func (c Config) IsModuleActive(name string) bool {
	for _, w := range c.Layout.Widgets {
		if w.Module == name {
			return true
		}
	}
	return false
}

func Default() Config {
	return Config{
		Server: ServerConfig{Port: 8080},
		Layout: LayoutConfig{
			MaxWidgets: 8,
			Direction:  "row",
			Widgets: []WidgetEntry{
				{Module: "clock"},
				{Module: "weather"},
			},
		},
		Modules: map[string]ModuleConfig{
			"clock": {
				Config: map[string]any{
					"format":      "24h",
					"showSeconds": true,
					"showDate":    true,
				},
			},
			"weather": {
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

	if len(cfg.Layout.Widgets) > cfg.Layout.MaxWidgets {
		return Config{}, fmt.Errorf("widgets count %d exceeds maxWidgets %d", len(cfg.Layout.Widgets), cfg.Layout.MaxWidgets)
	}

	return cfg, nil
}
