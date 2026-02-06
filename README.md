# news-go

`news-go` 是一个计划中的 Go 新闻聚合项目。

当前仓库还在 **0→1 初始化阶段**：已具备许可证与项目规划文档，但尚未进入可运行实现。

## 项目完成度（当前评估）

- **整体完成度：约 10%**（偏规划阶段，工程实现尚未开始）。
- **已完成**：
  - 基础开源信息（LICENSE）
  - Roadmap 分阶段任务拆解
  - 本次新增的完成度评估与优化清单
- **未完成（关键）**：
  - Go module 与可运行服务入口
  - 数据采集、存储、查询 API
  - 测试体系与 CI

> 详细评估请见 `PROJECT_REVIEW.md`。

## 建议优先补齐的能力（按优先级）

1. **工程骨架**：`go mod` + `cmd/api` + `internal/*` + `Makefile`
2. **最小可用 API**：`GET /healthz`、`GET /v1/articles`
3. **最小数据链路**：RSS 抓取 -> 去重 -> SQLite 落库
4. **质量门禁**：`go test ./...` + `go vet ./...` + CI
5. **可观测性**：结构化日志、健康检查、基础指标

## 快速启动建议（下一步直接做）

- 初始化 Go 工程：`go mod init news-go`
- 创建目录：
  - `cmd/api`
  - `internal/{app,config,crawler,news,storage,httpapi}`
- 建立最小运行闭环：
  - 启动 HTTP 服务
  - 提供 `/healthz`
  - 添加一条 `go test` 的示例测试

## 文档导航

- `ROADMAP.md`：分阶段实施路线与验收标准
- `PROJECT_REVIEW.md`：当前完成度评估、短板与优化方案

