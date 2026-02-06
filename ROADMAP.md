# Roadmap（执行版）

## Phase 0：项目初始化（已完成）

- [x] 初始化 Go module 与目录结构
- [x] Makefile / .gitignore / .env.example
- [x] 基础 README

## Phase 1：抓取 MVP（已完成）

- [x] RSS 抓取器（单源）
- [x] URL hash 去重
- [x] SQLite 入库（含内存回退）
- [x] `GET /v1/articles/{id}`

## Phase 2：查询与稳定性（进行中）

- [x] 列表分页
- [x] 关键词/来源/时间过滤
- [x] `readyz`
- [x] 结构化请求日志
- [x] 抓取定时同步 + 失败重试（基础）
- [ ] 指标暴露与告警
