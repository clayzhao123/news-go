# news-go

一个使用 Go 构建的新闻聚合服务（已完成 P0 最小可运行骨架）。

## 当前状态

已具备：
- Go 工程初始化（`go.mod`）
- HTTP 服务入口（`cmd/api`）
- 基础接口：`GET /healthz`、`GET /v1/articles`
- 内存版文章仓储（用于本地联调）
- SQLite 初始 schema（`db/schema.sql`）
- 基础测试与 CI（`go test` + `go vet`）

## 快速开始

```bash
cp .env.example .env
make run
```

默认监听 `:8080`。

## 常用命令

```bash
make run
make test
make vet
make fmt
```

## API

### 健康检查

```http
GET /healthz
```

示例响应：

```json
{"status":"ok"}
```

### 文章列表

```http
GET /v1/articles?limit=20&offset=0
```

## 目录结构

- `cmd/api`：服务启动入口
- `internal/app`：应用启动与 server 装配
- `internal/httpapi`：HTTP 路由与 handler
- `internal/news`：领域模型
- `internal/storage`：存储抽象与内存实现
- `internal/config`：配置加载
- `db/schema.sql`：SQLite 初始表结构

## Roadmap

后续阶段规划见 `ROADMAP.md`，完成度评估见 `PROJECT_REVIEW.md`。
