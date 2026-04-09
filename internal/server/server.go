package server

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/digibituk/resilver/internal/config"
)

type Server struct {
	cfg   config.Config
	webFS fs.FS
}

func New(cfg config.Config, webFS fs.FS) *Server {
	return &Server{cfg: cfg, webFS: webFS}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/config", s.handleConfig)
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
