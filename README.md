# news-go

一个使用 Go 构建的新闻聚合服务（P1 已完成，P2 已开始）。

## 当前能力

- 可运行 HTTP 服务：`cmd/api`
- 接口：
  - `GET /healthz`
  - `GET /readyz`
  - `GET /v1/articles?limit=&offset=&q=&source=&from=&to=`
  - `GET /v1/articles/{id}`
- RSS 抓取：启动抓取 + 定时同步（可配置间隔/重试）
- 存储：优先 SQLite，不可用时回退内存仓储
- 去重：`url_hash` 唯一约束 + upsert
- 结构化请求日志：包含 method/path/status/duration/request_id

## 快速开始

```bash
cp .env.example .env
make run
```

## 配置

- `HTTP_ADDR`：监听地址
- `DB_PATH`：SQLite 文件路径
- `RSS_FEED_URL`：RSS 源
- `RSS_SYNC_INTERVAL_SEC`：定时同步间隔（秒）
- `RSS_MAX_RETRIES`：单次同步失败重试次数

## 常用命令

```bash
make test
make vet
make fmt
```
