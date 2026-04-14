package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	resilver "github.com/digibituk/resilver"
	"github.com/digibituk/resilver/internal/config"
	"github.com/digibituk/resilver/internal/server"
	"github.com/digibituk/resilver/internal/update"
)

var version = "dev"

func main() {
	configPath := flag.String("config", "", "path to config file (uses embedded defaults if omitted)")
	port := flag.Int("port", 0, "override server port")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	log.Printf("resilver %s", version)

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if *port != 0 {
		cfg.Server.Port = *port
	}

	if cfg.Update.Enabled {
		if _, err := update.ParseVersion(version); err != nil {
			log.Printf("auto-update disabled: version %q is not a valid semver (build with ldflags to set version)", version)
		} else if cfg.Update.IntervalHours < 1 {
			log.Printf("auto-update disabled: intervalHours must be at least 1")
		} else {
			client := &update.Client{
				BaseURL:    "https://api.github.com",
				HTTPClient: &http.Client{Timeout: 30 * time.Second},
				Owner:      "digibituk",
				Repo:       "resilver",
			}
			updater := &update.Updater{
				CurrentVersion: version,
				GOOS:           runtime.GOOS,
				GOARCH:         runtime.GOARCH,
				Client:         client,
				Replacer:       update.ReplaceBinary,
				Restarter:      update.RestartSelf,
			}
			stop := make(chan struct{})
			defer close(stop)
			go updater.Run(time.Duration(cfg.Update.IntervalHours)*time.Hour, stop)
		}
	}

	if binPath, err := os.Executable(); err == nil {
		log.Printf("auto-update: cleaning up backup if present")
		update.CleanupBackup(binPath)
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
