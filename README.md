# news-go

一个使用 Go 构建的新闻聚合服务（P0 已完成，P1 进行中）。

## 当前能力

- 可运行 HTTP 服务：`cmd/api`
- 接口：
  - `GET /healthz`
  - `GET /v1/articles?limit=&offset=`
  - `GET /v1/articles/{id}`
- RSS 抓取：启动时从 `RSS_FEED_URL` 拉取并写入仓储
- 存储：默认 SQLite（失败时回退内存仓储）
- 去重：基于 `url_hash` 唯一约束 + upsert

## 快速开始

```bash
cp .env.example .env
make run
```

## 常用命令

```bash
make test
make vet
make fmt
```

## 配置

- `HTTP_ADDR`：监听地址（默认 `:8080`）
- `DB_PATH`：SQLite 文件路径
- `RSS_FEED_URL`：RSS 源地址

## 路线

- `ROADMAP.md`：里程碑与阶段任务
- `PROJECT_REVIEW.md`：当前完成度评估与后续优化
