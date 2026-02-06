# Roadmap（执行版）

## Phase 0：项目初始化（已完成）

- [x] 初始化 Go module 与目录结构
- [x] 增加 `Makefile`（`run`/`test`/`vet`）
- [x] 增加 `.gitignore`、`.env.example`
- [x] 完成基础 README（运行方式 + 目录说明）

## Phase 1：抓取 MVP（进行中）

- [x] 实现 RSS 抓取器（单源，启动时抓取）
- [x] 建立去重规则（URL hash + 唯一约束）
- [x] 使用 SQLite 完成入库（含回退内存仓储）
- [x] 提供 `GET /v1/articles/{id}`
- [ ] 增加多源抓取与调度

## Phase 2：查询 API

- [x] 提供文章列表 API（分页）
- [x] 默认按发布时间倒序
- [ ] 关键词/来源/时间过滤

## Phase 3：稳定性与工程化

- [x] 覆盖基础单测
- [x] 配置 CI（`go test` + `go vet`）
- [ ] 增加 `readyz`、结构化日志、指标
