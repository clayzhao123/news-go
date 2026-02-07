package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"news-go/internal/storage"
)

type Handler struct {
	repo storage.ArticleRepository
}

func NewHandler(repo storage.ArticleRepository) *Handler { return &Handler{repo: repo} }

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", h.healthz)
	mux.HandleFunc("/readyz", h.readyz)
	mux.HandleFunc("/v1/articles", h.listArticles)
	mux.HandleFunc("/v1/articles/", h.getArticleByID)
}

func (h *Handler) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) readyz(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.Ready(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready", "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (h *Handler) listArticles(w http.ResponseWriter, r *http.Request) {
	limit := clamp(parseIntOrDefault(r.URL.Query().Get("limit"), 20), 1, 100)
	offset := parseIntOrDefault(r.URL.Query().Get("offset"), 0)
	if offset < 0 {
		offset = 0
	}

	opts := storage.ListOptions{
		Limit:   limit,
		Offset:  offset,
		Keyword: strings.TrimSpace(r.URL.Query().Get("q")),
		Source:  strings.TrimSpace(r.URL.Query().Get("source")),
	}
	if from := strings.TrimSpace(r.URL.Query().Get("from")); from != "" {
		t, err := parseRFC3339Param(from)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid from, expected RFC3339"})
			return
		}
		opts.PublishedFrom = t
	}
	if to := strings.TrimSpace(r.URL.Query().Get("to")); to != "" {
		t, err := parseRFC3339Param(to)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid to, expected RFC3339"})
			return
		}
		opts.PublishedTo = t
	}
	if !opts.PublishedFrom.IsZero() && !opts.PublishedTo.IsZero() && opts.PublishedFrom.After(opts.PublishedTo) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid time range: from must be before or equal to to"})
		return
	}

	articles, err := h.repo.ListArticles(r.Context(), opts)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list articles"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": articles, "limit": limit, "offset": offset})
}

func (h *Handler) getArticleByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/v1/articles/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid article id"})
		return
	}
	article, err := h.repo.GetArticleByID(r.Context(), id)
	if err != nil {
		if err == storage.ErrNotFound {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "article not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get article"})
		return
	}
	writeJSON(w, http.StatusOK, article)
}

func parseIntOrDefault(v string, fallback int) int {
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func parseRFC3339Param(v string) (time.Time, error) {
	return time.Parse(time.RFC3339, v)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
