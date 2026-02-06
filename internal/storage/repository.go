package storage

import (
	"context"
	"sort"
	"sync"
	"time"

	"news-go/internal/news"
)

type ArticleRepository interface {
	ListArticles(ctx context.Context, limit, offset int) ([]news.Article, error)
}

type MemoryArticleRepository struct {
	mu       sync.RWMutex
	articles []news.Article
}

func NewMemoryArticleRepository() *MemoryArticleRepository {
	now := time.Now().UTC()
	return &MemoryArticleRepository{
		articles: []news.Article{
			{ID: 1, Title: "news-go initialized", URL: "https://example.com/news-go-init", Source: "system", PublishedAt: now},
			{ID: 2, Title: "first endpoint online", URL: "https://example.com/healthz", Source: "system", PublishedAt: now.Add(-1 * time.Hour)},
		},
	}
}

func (r *MemoryArticleRepository) ListArticles(_ context.Context, limit, offset int) ([]news.Article, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]news.Article, len(r.articles))
	copy(items, r.articles)

	sort.Slice(items, func(i, j int) bool {
		return items[i].PublishedAt.After(items[j].PublishedAt)
	})

	if offset >= len(items) {
		return []news.Article{}, nil
	}

	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end], nil
}
