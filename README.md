# news-go

这是一个**混合仓库**，目前包含两部分能力：

1. **Go 新闻聚合 API（主线）**：RSS 抓取、入库、去重、HTTP 查询接口。
2. **Python Streamlit 摘要 Demo（实验）**：按来源权重生成中文每日摘要页面。

> 如果你只想快速上手并对接接口，优先使用 **Go API**。

---

## 1) Go 新闻聚合 API（推荐）

### 功能

- HTTP 服务（`cmd/api`）
- 健康检查：`GET /healthz`
- 就绪检查：`GET /readyz`
- 文章列表：`GET /v1/articles?limit=&offset=&q=&source=&from=&to=`
  - `from/to` 必须是 RFC3339 时间格式
  - 且 `from <= to`，否则返回 `400`
- 文章详情：`GET /v1/articles/{id}`
- RSS 启动抓取 + 定时同步（可配置重试）
- 存储优先 SQLite，不可用时回退内存仓储
- URL 去重（`url_hash` 唯一约束 + upsert）

### 环境变量

可通过 `.env` 文件或系统环境变量传入：

- `APP_ENV`（默认 `dev`）
- `HTTP_ADDR`（默认 `:8080`）
- `DB_PATH`（默认 `./data/news.db`）
- `RSS_FEED_URL`（默认 `https://hnrss.org/frontpage`）
- `RSS_SYNC_INTERVAL_SEC`（默认 `300`）
- `RSS_MAX_RETRIES`（默认 `2`）

### 快速启动

```bash
cp .env.example .env
make run
```

启动后默认监听 `http://localhost:8080`。

### 常用命令

```bash
make test
make vet
make fmt
```

### 接口示例

```bash
# 健康检查
curl http://localhost:8080/healthz

# 查询最近文章
curl "http://localhost:8080/v1/articles?limit=10&offset=0"

# 按关键词和来源过滤
curl "http://localhost:8080/v1/articles?q=ai&source=Hacker%20News"
```

---

## 2) Python Streamlit 摘要 Demo（可选）

该部分位于：

- `app.py`
- `src/news_pipeline.py`
- `data/sources.json`

用于展示“每日新闻评分与摘要”界面，适合演示；与 Go API 目前不是同一运行时。

### 运行方式

```bash
python -m venv .venv
source .venv/bin/activate  # Windows: .venv\Scripts\activate
pip install -r requirements.txt
streamlit run app.py
```

浏览器打开 `http://localhost:8501`。

---

## 项目结构（核心）

```text
news-go/
├─ cmd/api/                  # Go 服务入口
├─ internal/                 # Go 业务逻辑（抓取/存储/API）
├─ db/schema.sql             # SQLite 初始化脚本
├─ app.py                    # Streamlit Demo 入口
├─ src/news_pipeline.py      # Python 摘要流程
├─ data/sources.json         # 新闻源配置
├─ .env.example
├─ Makefile
└─ README.md
```

---

## 说明

- 当前仓库处于“Go API 主线 + Python Demo 并存”状态。
- 若后续要统一架构，建议让 Streamlit 直接消费 Go API，避免双套抓取逻辑。
