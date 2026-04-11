package server

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/digibituk/resilver/internal/config"
	"github.com/digibituk/resilver/internal/news"
	"github.com/digibituk/resilver/internal/weather"
)

type Server struct {
	cfg           config.Config
	webFS         fs.FS
	weatherClient *weather.CachedClient
	newsClient    *news.CachedClient
}

func New(cfg config.Config, webFS fs.FS) *Server {
	return NewWithWeatherURL(cfg, webFS, "https://api.open-meteo.com")
}

func NewWithWeatherURL(cfg config.Config, webFS fs.FS, weatherURL string) *Server {
	weatherRefresh := 600
	if wCfg, ok := cfg.Modules["weather"]; ok {
		if v, ok := wCfg.Config["refreshIntervalSeconds"]; ok {
			if f, ok := v.(float64); ok {
				weatherRefresh = int(f)
			}
		}
	}

	newsRefresh := 1800
	if nCfg, ok := cfg.Modules["news"]; ok {
		if v, ok := nCfg.Config["refreshIntervalSeconds"]; ok {
			if f, ok := v.(float64); ok {
				newsRefresh = int(f)
			}
		}
	}

	return &Server{
		cfg:           cfg,
		webFS:         webFS,
		weatherClient: weather.NewCachedClient(weatherURL, time.Duration(weatherRefresh)*time.Second),
		newsClient:    news.NewCachedClient(time.Duration(newsRefresh) * time.Second),
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/weather", s.handleWeather)
	mux.HandleFunc("/api/news", s.handleNews)
	mux.Handle("/", http.FileServer(http.FS(s.webFS)))

	return mux
}

func (s *Server) ListenAndServe() error {
	addr := fmt.Sprintf(":%d", s.cfg.Server.Port)
	log.Printf("resilver listening on %s", addr)
	return http.ListenAndServe(addr, s.Handler())
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.cfg)
}

func (s *Server) handleWeather(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.IsModuleActive("weather") {
		http.Error(w, "weather module not enabled", http.StatusNotFound)
		return
	}

	wCfg := s.cfg.Modules["weather"]

	lat, _ := wCfg.Config["latitude"].(float64)
	lon, _ := wCfg.Config["longitude"].(float64)
	units, _ := wCfg.Config["units"].(string)
	if units == "" {
		units = "celsius"
	}

	data, err := s.weatherClient.Fetch(lat, lon, units)
	if err != nil {
		log.Printf("weather fetch error: %v", err)
		http.Error(w, "failed to fetch weather data", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *Server) handleNews(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.IsModuleActive("news") {
		http.Error(w, "news module not enabled", http.StatusNotFound)
		return
	}

	nCfg := s.cfg.Modules["news"]

	var feedURLs []string
	if urls, ok := nCfg.Config["feedUrls"].([]any); ok {
		for _, u := range urls {
			if s, ok := u.(string); ok {
				feedURLs = append(feedURLs, s)
			}
		}
	}

	maxItems := 5
	if v, ok := nCfg.Config["maxItems"].(float64); ok {
		maxItems = int(v)
	}

	items, err := s.newsClient.Fetch(feedURLs, maxItems)
	if err != nil {
		log.Printf("news fetch error: %v", err)
		http.Error(w, "failed to fetch news data", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}
