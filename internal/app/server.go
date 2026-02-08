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
	syncer := newRSSSyncer(cfg, repo)
	syncer.start(context.Background())

	h := httpapi.NewHandler(repo)
	mux := http.NewServeMux()
	h.Register(mux)

	return &http.Server{Addr: cfg.HTTPAddr, Handler: httpapi.LoggingMiddleware(mux)}
}

func buildRepository(cfg config.Config) storage.ArticleRepository {
	repo, err := storage.NewSQLiteArticleRepository(cfg.DBPath, "db/schema.sql")
	if err != nil {
		log.Printf("sqlite init failed, fallback to memory repo: %v", err)
		return storage.NewMemoryArticleRepository()
	}
	return repo
}

type rssSyncer struct {
	cfg     config.Config
	repo    storage.ArticleRepository
	fetcher *crawler.RSSFetcher
}

func newRSSSyncer(cfg config.Config, repo storage.ArticleRepository) *rssSyncer {
	return &rssSyncer{cfg: cfg, repo: repo, fetcher: crawler.NewRSSFetcher(10 * time.Second)}
}

func (s *rssSyncer) start(ctx context.Context) {
	go s.syncWithRetry(ctx)
	if s.cfg.RSSSyncIntervalSec <= 0 {
		return
	}
	go func() {
		ticker := time.NewTicker(time.Duration(s.cfg.RSSSyncIntervalSec) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.syncWithRetry(ctx)
			}
		}
	}()
}

func (s *rssSyncer) syncWithRetry(ctx context.Context) {
	attempts := s.cfg.RSSMaxRetries + 1
	if attempts < 1 {
		attempts = 1
	}
	for i := 1; i <= attempts; i++ {
		if err := s.syncOnce(ctx); err != nil {
			log.Printf("rss sync attempt=%d/%d failed: %v", i, attempts, err)
			if i < attempts {
				time.Sleep(2 * time.Second)
			}
			continue
		}
		return
	}
}

func (s *rssSyncer) syncOnce(ctx context.Context) error {
	callCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	items, err := s.fetcher.Fetch(callCtx, s.cfg.RSSFeedURL, s.cfg.RSSUserAgent)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}
	if err := s.repo.UpsertArticles(callCtx, items); err != nil {
		return err
	}
	log.Printf("event=rss_sync status=ok fetched=%d", len(items))
	return nil
}

func Run(cfg config.Config) error {
	srv := NewServer(cfg)
	fmt.Printf("news-go listening on %s\n", cfg.HTTPAddr)
	return srv.ListenAndServe()
}
