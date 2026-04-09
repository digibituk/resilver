package main

import (
	"flag"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	resilver "github.com/digibituk/resilver"
	"github.com/digibituk/resilver/internal/config"
	"github.com/digibituk/resilver/internal/server"
)

var version = "dev"

func main() {
	configPath := flag.String("config", defaultConfigPath(), "path to config file")
	port := flag.Int("port", 0, "override server port")
	flag.Parse()

	log.Printf("resilver %s", version)

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if *port != 0 {
		cfg.Server.Port = *port
	}

	webRoot, err := fs.Sub(resilver.WebFS, "web")
	if err != nil {
		log.Fatalf("failed to access embedded web assets: %v", err)
	}

	srv := server.New(cfg, webRoot)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func defaultConfigPath() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "resilver", "config.json")
	}
	return "config.json"
}
