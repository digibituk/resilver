package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	resilver "github.com/digibituk/resilver"
)

type Config struct {
	Server  ServerConfig            `json:"server"`
	Layout  LayoutConfig            `json:"layout"`
	Modules map[string]ModuleConfig `json:"modules"`
	Update  UpdateConfig            `json:"update"`
}

type UpdateConfig struct {
	Enabled       bool `json:"enabled"`
	IntervalHours int  `json:"intervalHours"`
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

// Load reads config from the given path. If path is empty or the file does
// not exist, it falls back to the embedded default config.json.
func Load(path string) (Config, error) {
	var data []byte

	if path != "" {
		raw, err := os.ReadFile(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				data = resilver.DefaultConfig
			} else {
				return Config{}, err
			}
		} else {
			data = raw
		}
	} else {
		data = resilver.DefaultConfig
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
