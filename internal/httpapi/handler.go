package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
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
	mux.HandleFunc("/v1/digest", h.dailyDigest)
	mux.HandleFunc("/", h.home)
}

func (h *Handler) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(homeHTML))
}

const homeHTML = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>news-go 可视化看板</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; margin: 0; background: #f6f8fa; color: #111; }
    .wrap { max-width: 900px; margin: 0 auto; padding: 24px; }
    h1 { margin: 0 0 8px; }
    .hint { color: #666; margin-bottom: 16px; }
    .toolbar { display: flex; gap: 8px; margin-bottom: 14px; }
    input { flex: 1; padding: 10px; border: 1px solid #ddd; border-radius: 8px; }
    button { padding: 10px 14px; border: 0; border-radius: 8px; background: #111827; color: #fff; cursor: pointer; }
    .card { background: white; border: 1px solid #e5e7eb; border-radius: 10px; padding: 12px 14px; margin-bottom: 10px; }
    .meta { color: #6b7280; font-size: 12px; margin-bottom: 6px; }
    a { color: #2563eb; text-decoration: none; }
  </style>
</head>
<body>
  <div class="wrap">
    <h1>news-go 可视化看板</h1>
    <p class="hint">优先展示策略版每日摘要（/v1/digest）。若尚未生成，则回退展示普通新闻列表。</p>
    <div class="toolbar">
      <input id="q" placeholder="输入关键词，比如 AI" />
      <button onclick="loadArticles()">刷新</button>
    </div>
    <div id="status" class="hint">加载中...</div>
    <div id="list"></div>
  </div>
<script>
async function loadArticles() {
  const q = document.getElementById('q').value.trim();
  const digestURL = '/v1/digest';
  const url = '/v1/articles?limit=20&offset=0' + (q ? ('&q=' + encodeURIComponent(q)) : '');
  const status = document.getElementById('status');
  const list = document.getElementById('list');
  status.textContent = '加载中...';
  list.innerHTML = '';
  try {
    let data = null;
    let items = [];
    if (!q) {
      const digestRes = await fetch(digestURL);
      if (digestRes.ok) {
        data = await digestRes.json();
        items = data.items || [];
        status.textContent = '策略摘要，共 ' + items.length + ' 条';
      }
    }
    if (!items.length) {
      const res = await fetch(url);
      data = await res.json();
      items = data.items || [];
      status.textContent = '普通列表，共 ' + items.length + ' 条';
    }
    if (!items.length) {
      list.innerHTML = '<div class="card">暂无数据（可能是 RSS 源暂时不可访问）</div>';
      return;
    }
    list.innerHTML = items.map(function (x) {
      return '<div class="card">'
        + '<div class="meta">#' + (x.id || '-') + ' · ' + (x.source || 'rss') + ' · ' + (x.published_at || '') + '</div>'
        + '<div><a href="' + (x.url || '#') + '" target="_blank">' + (x.title || '(无标题)') + '</a></div>'
        + '</div>';
    }).join('');
  } catch (e) {
    status.textContent = '加载失败';
    list.innerHTML = '<div class="card">请求失败，请检查服务日志。</div>';
  }
}
loadArticles();
</script>
</body>
</html>`

const dailyDigestPath = "data/daily_digest.json"

func (h *Handler) dailyDigest(w http.ResponseWriter, _ *http.Request) {
	body, err := os.ReadFile(dailyDigestPath)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "daily digest not generated",
			"hint":  "run: python -m src.digest_job",
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
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
