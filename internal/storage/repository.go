package storage

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"news-go/internal/news"
)

var ErrNotFound = errors.New("not found")

type ArticleRepository interface {
	ListArticles(ctx context.Context, limit, offset int) ([]news.Article, error)
	GetArticleByID(ctx context.Context, id int64) (news.Article, error)
	UpsertArticles(ctx context.Context, articles []news.Article) error
}

type MemoryArticleRepository struct {
	mu       sync.RWMutex
	articles []news.Article
}

func NewMemoryArticleRepository() *MemoryArticleRepository {
	now := time.Now().UTC()
	return &MemoryArticleRepository{articles: []news.Article{{ID: 1, Title: "news-go initialized", URL: "https://example.com/news-go-init", Source: "system", PublishedAt: now}, {ID: 2, Title: "first endpoint online", URL: "https://example.com/healthz", Source: "system", PublishedAt: now.Add(-1 * time.Hour)}}}
}

func (r *MemoryArticleRepository) ListArticles(_ context.Context, limit, offset int) ([]news.Article, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := append([]news.Article(nil), r.articles...)
	sort.Slice(items, func(i, j int) bool { return items[i].PublishedAt.After(items[j].PublishedAt) })
	if offset >= len(items) {
		return []news.Article{}, nil
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end], nil
}

func (r *MemoryArticleRepository) GetArticleByID(_ context.Context, id int64) (news.Article, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, a := range r.articles {
		if a.ID == id {
			return a, nil
		}
	}
	return news.Article{}, ErrNotFound
}

func (r *MemoryArticleRepository) UpsertArticles(_ context.Context, articles []news.Article) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	byURL := map[string]news.Article{}
	var maxID int64
	for _, a := range r.articles {
		byURL[a.URL] = a
		if a.ID > maxID {
			maxID = a.ID
		}
	}
	for _, a := range articles {
		if old, ok := byURL[a.URL]; ok {
			a.ID = old.ID
		} else {
			maxID++
			a.ID = maxID
		}
		byURL[a.URL] = a
	}
	r.articles = r.articles[:0]
	for _, a := range byURL {
		r.articles = append(r.articles, a)
	}
	return nil
}

type SQLiteArticleRepository struct{ dbPath string }

func NewSQLiteArticleRepository(dbPath, schemaPath string) (*SQLiteArticleRepository, error) {
	if _, err := exec.LookPath("sqlite3"); err != nil {
		return nil, fmt.Errorf("sqlite3 binary not found")
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, err
	}
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, err
	}
	if _, err := runSQLite(dbPath, string(schema)); err != nil {
		return nil, err
	}
	return &SQLiteArticleRepository{dbPath: dbPath}, nil
}

func (r *SQLiteArticleRepository) ListArticles(_ context.Context, limit, offset int) ([]news.Article, error) {
	q := fmt.Sprintf("SELECT id, title, url, COALESCE(content,''), COALESCE(published_at,''), 'rss' FROM articles ORDER BY published_at DESC LIMIT %d OFFSET %d;", limit, offset)
	out, err := runSQLite(r.dbPath, q)
	if err != nil {
		return nil, err
	}
	return parseRows(out), nil
}

func (r *SQLiteArticleRepository) GetArticleByID(_ context.Context, id int64) (news.Article, error) {
	q := fmt.Sprintf("SELECT id, title, url, COALESCE(content,''), COALESCE(published_at,''), 'rss' FROM articles WHERE id = %d;", id)
	out, err := runSQLite(r.dbPath, q)
	if err != nil {
		return news.Article{}, err
	}
	items := parseRows(out)
	if len(items) == 0 {
		return news.Article{}, ErrNotFound
	}
	return items[0], nil
}

func (r *SQLiteArticleRepository) UpsertArticles(_ context.Context, articles []news.Article) error {
	var b strings.Builder
	for _, a := range articles {
		published := a.PublishedAt.UTC().Format(time.RFC3339)
		b.WriteString(fmt.Sprintf("INSERT INTO articles (title, url, url_hash, content, published_at) VALUES ('%s','%s','%s','%s','%s') ON CONFLICT(url_hash) DO UPDATE SET title=excluded.title, content=excluded.content, published_at=excluded.published_at;", esc(a.Title), esc(a.URL), hashURL(a.URL), esc(a.Content), published))
	}
	_, err := runSQLite(r.dbPath, b.String())
	return err
}

func runSQLite(path, query string) (string, error) {
	cmd := exec.Command("sqlite3", "-separator", "|", path, query)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("sqlite error: %v, output: %s", err, string(out))
	}
	return string(out), nil
}

func parseRows(raw string) []news.Article {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []news.Article{}
	}
	items := make([]news.Article, 0, len(lines))
	for _, line := range lines {
		cols := strings.Split(line, "|")
		if len(cols) < 6 {
			continue
		}
		id, _ := strconv.ParseInt(cols[0], 10, 64)
		t, _ := time.Parse(time.RFC3339, cols[4])
		items = append(items, news.Article{ID: id, Title: cols[1], URL: cols[2], Content: cols[3], PublishedAt: t, Source: cols[5]})
	}
	return items
}

func hashURL(url string) string {
	sum := sha256.Sum256([]byte(url))
	return fmt.Sprintf("%x", sum)
}

func esc(v string) string { return strings.ReplaceAll(v, "'", "''") }
