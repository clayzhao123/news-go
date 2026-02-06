package storage

import (
	"context"
	"testing"
	"time"

	"news-go/internal/news"
)

func TestMemoryUpsertDeduplicate(t *testing.T) {
	repo := NewMemoryArticleRepository()
	ctx := context.Background()
	now := time.Now().UTC()
	input := []news.Article{
		{Title: "A", URL: "https://example.com/1", Content: "v1", PublishedAt: now},
		{Title: "A updated", URL: "https://example.com/1", Content: "v2", PublishedAt: now.Add(time.Minute)},
	}
	if err := repo.UpsertArticles(ctx, input); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	items, err := repo.ListArticles(ctx, 100, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	count := 0
	for _, it := range items {
		if it.URL == "https://example.com/1" {
			count++
			if it.Title != "A updated" {
				t.Fatalf("expected updated title, got %q", it.Title)
			}
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one deduped row, got %d", count)
	}
}
