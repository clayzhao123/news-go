package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"news-go/internal/config"
	"news-go/internal/crawler"
	"news-go/internal/httpapi"
	"news-go/internal/storage"
)

func NewServer(cfg config.Config) *http.Server {
	repo := buildRepository(cfg)
	seedFromRSS(cfg, repo)

	h := httpapi.NewHandler(repo)
	mux := http.NewServeMux()
	h.Register(mux)

	return &http.Server{Addr: cfg.HTTPAddr, Handler: mux}
}

func buildRepository(cfg config.Config) storage.ArticleRepository {
	repo, err := storage.NewSQLiteArticleRepository(cfg.DBPath, "db/schema.sql")
	if err != nil {
		log.Printf("sqlite init failed, fallback to memory repo: %v", err)
		return storage.NewMemoryArticleRepository()
	}
	return repo
}

func seedFromRSS(cfg config.Config, repo storage.ArticleRepository) {
	fetcher := crawler.NewRSSFetcher(10 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	items, err := fetcher.Fetch(ctx, cfg.RSSFeedURL)
	if err != nil {
		log.Printf("rss fetch failed: %v", err)
		return
	}
	if len(items) == 0 {
		return
	}
	if err := repo.UpsertArticles(ctx, items); err != nil {
		log.Printf("rss upsert failed: %v", err)
	}
}

func Run(cfg config.Config) error {
	srv := NewServer(cfg)
	fmt.Printf("news-go listening on %s\n", cfg.HTTPAddr)
	return srv.ListenAndServe()
}
