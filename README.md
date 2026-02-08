# news-go

`news-go` 是一个新闻聚合项目：

- **Go API（主线）**：抓取 RSS、存储、查询接口。
- **Python Streamlit（可选）**：摘要演示页面。

---

## 1. 先看结论（避免你反复踩坑）

1. 你在 Windows 报 `make` 找不到，是因为系统通常默认**没安装 make**，不是项目坏了。
2. Windows 上直接跑：`go run ./cmd/api`（不需要 make）。
3. 如果日志出现 `sqlite3 not installed, using in-memory repository`，服务仍然能跑，只是数据不落盘。
4. `/healthz` 返回 `{"status":"ok"}` 就表示服务已跑通。

---

## 2. 快速启动（Windows PowerShell，推荐）

### 步骤 1：进入目录

```powershell
cd news-go
```

### 步骤 2：复制配置

```powershell
Copy-Item .env.example .env
```

### 步骤 3：启动服务（不使用 make）

```powershell
go run ./cmd/api
```

看到类似输出表示成功：

```text
news-go listening on :8080
```

### 步骤 4：验证

新开一个 PowerShell：

```powershell
curl http://localhost:8080/healthz
```

看到：

```json
{"status":"ok"}
```

---

## 3. 常见问题

### Q1：`make run` 报错（make 未识别）

原因：你机器没装 make。  
解决：直接执行以下等价命令：

```powershell
go run ./cmd/api
```

### Q2：出现 `sqlite3 not installed, using in-memory repository`

原因：没装 `sqlite3` 命令行工具。  
影响：服务可用，但重启后数据会丢失（内存模式）。

可选安装（Windows）：

```powershell
winget install SQLite.SQLite
```

安装后重开终端，再执行 `go run ./cmd/api`。

### Q3：服务起来了，但文章为空

常见原因：默认 RSS 源在部分网络环境会 403。

可尝试修改 `.env`：

```env
RSS_MAX_RETRIES=0
RSS_FEED_URL=<你能访问的 RSS 地址>
RSS_USER_AGENT=news-go/1.0
```

---

## 4. 其他系统启动方式（macOS/Linux）

```bash
cd news-go
cp .env.example .env
```

### 4) 启动服务

```bash
make run
```

如果你没有 make，也可以直接：

```bash
go run ./cmd/api
```

---

## 5. 接口示例

```bash
# 健康检查
curl http://localhost:8080/healthz

# 就绪检查
curl http://localhost:8080/readyz

# 文章列表
curl "http://localhost:8080/v1/articles?limit=10&offset=0"

# 关键词 + 来源过滤
curl "http://localhost:8080/v1/articles?q=ai&source=Hacker%20News"
```

参数说明：

- `limit`：每页数量
- `offset`：从第几条开始
- `q`：关键词
- `source`：来源名
- `from/to`：时间范围（必须是 RFC3339，如 `2026-01-01T00:00:00Z`）

---

## F. 开发常用命令

> Windows 若未安装 make，请用右侧 Go 原生命令。

```bash
make test   # 或 go test ./...
make vet    # 或 go vet ./...
make fmt    # 或 gofmt -w ./cmd ./internal
```

---

## G. 可选：运行 Python Streamlit 摘要 Demo

这部分是演示页面，和 Go API 不是同一运行时。

- `limit`：每页数量
- `offset`：从第几条开始
- `q`：关键词
- `source`：来源名
- `from/to`：时间范围（必须是 RFC3339，如 `2026-01-01T00:00:00Z`）

---

## F. 开发常用命令

> Windows 若未安装 make，请用右侧 Go 原生命令。

```bash
make test   # 或 go test ./...
make vet    # 或 go vet ./...
make fmt    # 或 gofmt -w ./cmd ./internal
```

---

## 6. 开发命令

```bash
make test   # 或 go test ./...
make vet    # 或 go vet ./...
make fmt    # 或 gofmt -w ./cmd ./internal
```

---

## 7. 可选：运行 Python Streamlit Demo

```bash
python -m venv .venv
source .venv/bin/activate  # Windows 用: .venv\Scripts\activate
pip install -r requirements.txt
streamlit run app.py
```

浏览器打开：`http://localhost:8501`。
