package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"news-go/internal/storage"
)

type Handler struct {
	repo storage.ArticleRepository
}

func NewHandler(repo storage.ArticleRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", h.healthz)
	mux.HandleFunc("/v1/articles", h.listArticles)
}

func (h *Handler) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) listArticles(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit := parseIntOrDefault(q.Get("limit"), 20)
	offset := parseIntOrDefault(q.Get("offset"), 0)

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	articles, err := h.repo.ListArticles(r.Context(), limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list articles"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items":  articles,
		"limit":  limit,
		"offset": offset,
	})
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

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
