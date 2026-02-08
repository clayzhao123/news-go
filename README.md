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
## B. macOS/Linux（可用 make）
## 先说结论（你最关心的）

- 这个仓库按 README 是可以跑起来的。  
- 但默认 RSS 源（`https://hnrss.org/frontpage`）在某些网络环境会返回 `403 Forbidden`，导致：
  - 首次启动会多等几秒（重试）；
  - 服务能启动，但文章列表可能为空。  
- 如果你只想先确认“服务活着”，可以直接访问 `/healthz`，不依赖 RSS 成功。
- 服务已改为**先启动 HTTP，再后台同步 RSS**，避免因 RSS 不可达阻塞启动。

---

## A. 电脑小白版：3 分钟跑通 Go API（推荐）

### 1) 你需要先安装什么

- Git（用于拉代码）
- Go 1.22+（用于运行后端）

不知道是否安装成功？在终端执行：

```bash
git --version
go version
```

只要能输出版本号就行。

### 2) 进入项目目录

```bash
cd news-go
```

### 3) 准备配置文件（直接复制模板）

```bash
cd news-go
cp .env.example .env
```

### 4) 启动服务

```bash
make run
```

如果你在 macOS/Linux 上也没有 `make`，同样可以直接用：

```bash
go run ./cmd/api
```

---

## C. 你原来为什么会失败（快速解释）


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
你会看到类似输出：

```text
news-go listening on :8080
```

> 提示：如果看到 RSS `403` 日志，不等于启动失败，只是拉新闻失败。

### 5) 打开另一个终端，验证是否跑通

```bash
curl http://localhost:8080/healthz
```

看到：

```json
{"status":"ok"}
```

就说明服务已经跑起来了 ✅

### 6) 再试一个实际接口

```bash
curl "http://localhost:8080/v1/articles?limit=10&offset=0"
```

如果返回空数组（`"items":[]`）通常是 RSS 源未抓到数据，不影响 API 框架本身运行。

---

## B. 如果文章一直为空：一键排查

你可以先把重试次数调低，减少等待时间：

打开 `.env`，把这行改成：

```env
RSS_MAX_RETRIES=0
```

然后重启 `make run`。

如果你希望更容易抓到数据，可以把 `.env` 中的 `RSS_FEED_URL` 改成你可访问的 RSS 地址。

另外可尝试设置请求头（部分站点会校验）：

```env
RSS_USER_AGENT=news-go/1.0
```

---

## C. 常用接口（复制就能用）

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

## D. 开发常用命令（不会写代码也可忽略）

```bash
make test
make vet
make fmt
```

---

## E. 可选：运行 Python Streamlit 摘要 Demo

这部分是演示页面，和 Go API 不是同一运行时。

```bash
python -m venv .venv
source .venv/bin/activate  # Windows 用: .venv\Scripts\activate
pip install -r requirements.txt
streamlit run app.py
```

浏览器打开：`http://localhost:8501`

---

## F. 项目结构（你可以先不懂，知道位置即可）

```text
news-go/
├─ cmd/api/                  # Go 服务入口
├─ internal/                 # Go 业务逻辑（抓取/存储/API）
├─ db/schema.sql             # SQLite 初始化脚本
├─ app.py                    # Streamlit Demo 入口
├─ src/news_pipeline.py      # Python 摘要流程
├─ data/sources.json         # 新闻源配置
├─ .env.example              # 环境变量模板
├─ Makefile
└─ README.md
```

```bash
python -m venv .venv
source .venv/bin/activate  # Windows 用: .venv\Scripts\activate
pip install -r requirements.txt
streamlit run app.py
```

浏览器打开：`http://localhost:8501`

浏览器打开：`http://localhost:8501`
## G. 给你一个最短“成功路径”

如果你只想确认项目没坏：

1. `cp .env.example .env`
2. `make run`
3. 新终端执行 `curl http://localhost:8080/healthz`
4. 看到 `{"status":"ok"}` 就算跑通

