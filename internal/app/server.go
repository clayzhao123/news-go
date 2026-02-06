package app

import (
	"fmt"
	"net/http"

	"news-go/internal/config"
	"news-go/internal/httpapi"
	"news-go/internal/storage"
)

func NewServer(cfg config.Config) *http.Server {
	repo := storage.NewMemoryArticleRepository()
	h := httpapi.NewHandler(repo)
	mux := http.NewServeMux()
	h.Register(mux)

	return &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: mux,
	}
}

func Run(cfg config.Config) error {
	srv := NewServer(cfg)
	fmt.Printf("news-go listening on %s\n", cfg.HTTPAddr)
	return srv.ListenAndServe()
}
