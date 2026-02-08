# news-go

一个给新手也能跑起来的新闻聚合项目（Go 主线 + Python 演示）。

- **主线推荐**：Go API（更完整、可用于接口联调）
- **可选演示**：Python Streamlit 页面（摘要 Demo）

---

## 先说你这次报错的根因

你截图里的错误是：

> `make : 无法将“make”项识别为 cmdlet...`

这说明 **Windows 上没有安装 `make`**（这是正常的，很多 Windows 机器默认都没有）。

所以并不是 Go 项目坏了，而是你执行了一个本机没有的命令。  
解决方式：

- 直接用 Go 原生命令（最简单）：`go run ./cmd/api`
- 或者安装 `make`（可选，不是必须）

---

## A. 小白可直接跑通（Windows PowerShell 版本）

> 你现在这个环境最推荐看这一段。

### 1) 进入项目目录

```powershell
cd news-go
```

### 2) 复制配置文件

```powershell
Copy-Item .env.example .env
```

### 3) 启动服务（不使用 make）

```powershell
go run ./cmd/api
```

看到类似下面内容就表示服务启动成功：

```text
news-go listening on :8080
```


如果看到这条日志也不用慌：

```text
sqlite3 not installed, using in-memory repository
```

它表示你的机器没装 `sqlite3` 命令行工具，程序会自动切到内存存储（功能可用，但重启后数据不保留）。

### 4) 新开一个 PowerShell 窗口验证

```powershell
curl http://localhost:8080/healthz
```

看到：

```json
{"status":"ok"}
```

就算跑通 ✅

---

### Windows 想使用本地 SQLite 持久化（可选）

默认不装 `sqlite3` 也能跑；如果你想把数据落盘，可安装 sqlite3 后再启动。

```powershell
winget install SQLite.SQLite
# 安装后重开终端，再执行
go run ./cmd/api
```

## B. macOS/Linux（可用 make）

```bash
cd news-go
cp .env.example .env
make run
```

如果你在 macOS/Linux 上也没有 `make`，同样可以直接用：

```bash
go run ./cmd/api
```

---

## C. 你原来为什么会失败（快速解释）

`make run` 实际只是 Makefile 的一个快捷方式，本质执行的是：

```bash
go run ./cmd/api
```

也就是说：

- `make` 没装 → `make run` 报错
- 但只要 Go 在，就可以直接 `go run ./cmd/api`

---

## D. 如果服务启动了但文章为空

默认 RSS 源（`https://hnrss.org/frontpage`）在某些网络环境会 `403 Forbidden`，这会导致文章列表为空，但服务本身是正常的。

可选排查：

1. 把 `.env` 里的重试调低（启动更快看到结果）：

```env
RSS_MAX_RETRIES=0
```

2. 换你能访问的 RSS：

```env
RSS_FEED_URL=<你的RSS地址>
```

3. 调整请求头（部分站点会校验）：

```env
RSS_USER_AGENT=news-go/1.0
```

服务已经改为**先启动 HTTP，再后台同步 RSS**，所以就算 RSS 抓取失败，`/healthz` 也应该可用。

---

## E. 常用接口（复制就能用）

```bash
# 健康检查
curl http://localhost:8080/healthz

# 就绪检查
curl http://localhost:8080/readyz

# 文章列表
curl "http://localhost:8080/v1/articles?limit=10&offset=0"

# 按关键词 + 来源过滤
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

```bash
python -m venv .venv
source .venv/bin/activate  # Windows 用: .venv\Scripts\activate
pip install -r requirements.txt
streamlit run app.py
```

浏览器打开：`http://localhost:8501`

