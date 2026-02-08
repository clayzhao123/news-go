package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"news-go/internal/news"
	"news-go/internal/storage"
)

type stubRepo struct {
	items    []news.Article
	readyErr error
}

func (s stubRepo) ListArticles(_ context.Context, opts storage.ListOptions) ([]news.Article, error) {
	items := make([]news.Article, 0, len(s.items))
	for _, it := range s.items {
		if opts.Keyword != "" && it.Title != opts.Keyword {
			continue
		}
		items = append(items, it)
	}
	if opts.Offset >= len(items) {
		return []news.Article{}, nil
	}
	end := opts.Offset + opts.Limit
	if end > len(items) {
		end = len(items)
	}
	return items[opts.Offset:end], nil
}

func (s stubRepo) GetArticleByID(_ context.Context, id int64) (news.Article, error) {
	for _, it := range s.items {
		if it.ID == id {
			return it, nil
		}
	}
	return news.Article{}, storage.ErrNotFound
}

func (s stubRepo) UpsertArticles(_ context.Context, _ []news.Article) error { return nil }
func (s stubRepo) Ready(_ context.Context) error                            { return s.readyErr }

func TestHomePage(t *testing.T) {
	h := NewHandler(stubRepo{})
	mux := http.NewServeMux()
	h.Register(mux)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("expected text/html content-type, got %q", ct)
	}
}

func TestDailyDigestNotFound(t *testing.T) {
	_ = os.Remove("data/daily_digest.json")
	h := NewHandler(stubRepo{})
	mux := http.NewServeMux()
	h.Register(mux)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/digest", nil))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestDailyDigestOK(t *testing.T) {
	if err := os.MkdirAll("data", 0o755); err != nil {
		t.Fatalf("mkdir data: %v", err)
	}
	content := []byte(`{"items":[{"title":"x"}]}`)
	if err := os.WriteFile("data/daily_digest.json", content, 0o644); err != nil {
		t.Fatalf("write digest: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove("data/daily_digest.json") })

	h := NewHandler(stubRepo{})
	mux := http.NewServeMux()
	h.Register(mux)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/digest", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if rr.Body.String() == "" {
		t.Fatalf("expected digest body")
	}
}

func TestHealthz(t *testing.T) {
	h := NewHandler(stubRepo{})
	mux := http.NewServeMux()
	h.Register(mux)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestReadyz(t *testing.T) {
	t.Run("ready", func(t *testing.T) {
		h := NewHandler(stubRepo{})
		mux := http.NewServeMux()
		h.Register(mux)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}
	})
	t.Run("not ready", func(t *testing.T) {
		h := NewHandler(stubRepo{readyErr: errors.New("down")})
		mux := http.NewServeMux()
		h.Register(mux)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		if rr.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected 503, got %d", rr.Code)
		}
	})
}

func TestListArticles(t *testing.T) {
	now := time.Now().UTC()
	h := NewHandler(stubRepo{items: []news.Article{{ID: 1, Title: "a", PublishedAt: now}, {ID: 2, Title: "b", PublishedAt: now}}})
	mux := http.NewServeMux()
	h.Register(mux)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/articles?limit=10&offset=0&q=a", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var body struct {
		Items []news.Article `json:"items"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil || len(body.Items) != 1 {
		t.Fatalf("unexpected response: %v len=%d", err, len(body.Items))
	}
}

func TestListArticlesInvalidTimeFilter(t *testing.T) {
	h := NewHandler(stubRepo{})
	mux := http.NewServeMux()
	h.Register(mux)

	t.Run("invalid from", func(t *testing.T) {
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/articles?from=bad-time", nil))
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
	})

	t.Run("invalid to", func(t *testing.T) {
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/articles?to=bad-time", nil))
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
	})
}

func TestListArticlesInvalidTimeRange(t *testing.T) {
	h := NewHandler(stubRepo{})
	mux := http.NewServeMux()
	h.Register(mux)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/articles?from=2025-01-03T00:00:00Z&to=2025-01-02T00:00:00Z", nil))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestGetArticleByID(t *testing.T) {
	now := time.Now().UTC()
	h := NewHandler(stubRepo{items: []news.Article{{ID: 7, Title: "detail", PublishedAt: now}}})
	mux := http.NewServeMux()
	h.Register(mux)

	t.Run("ok", func(t *testing.T) {
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/articles/7", nil))
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/articles/9", nil))
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rr.Code)
		}
	})
}
