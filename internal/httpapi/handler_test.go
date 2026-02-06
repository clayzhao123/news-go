package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"news-go/internal/news"
)

type stubRepo struct {
	items []news.Article
}

func (s stubRepo) ListArticles(_ context.Context, limit, offset int) ([]news.Article, error) {
	if offset >= len(s.items) {
		return []news.Article{}, nil
	}
	end := offset + limit
	if end > len(s.items) {
		end = len(s.items)
	}
	return s.items[offset:end], nil
}

func TestHealthz(t *testing.T) {
	h := NewHandler(stubRepo{})
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestListArticles(t *testing.T) {
	now := time.Now().UTC()
	h := NewHandler(stubRepo{items: []news.Article{{ID: 1, Title: "a", PublishedAt: now}}})
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/v1/articles?limit=10&offset=0", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var body struct {
		Items []news.Article `json:"items"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(body.Items) != 1 {
		t.Fatalf("expected 1 article, got %d", len(body.Items))
	}
}
