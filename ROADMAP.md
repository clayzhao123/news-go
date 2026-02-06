# Roadmap（执行版）

## Phase 0：项目初始化（1~2 天）

- [ ] 初始化 Go module 与目录结构
- [ ] 增加 `Makefile`（`run`/`test`/`vet`）
- [ ] 增加 `.gitignore`、`.env.example`
- [ ] 完成基础 README（运行方式 + 目录说明）

**阶段产出**：项目可启动空服务并通过本地基础检查。

## Phase 1：抓取 MVP（3~5 天）

- [ ] 实现 RSS 抓取器
- [ ] 定义文章实体（title/url/source/published_at/content）
- [ ] 建立去重规则（URL hash + 唯一约束）
- [ ] 使用 SQLite 完成入库

**阶段产出**：可从至少一个源抓取并落库。

## Phase 2：查询 API（2~4 天）

- [ ] 提供文章列表 API（分页/关键词/来源过滤）
- [ ] 提供文章详情 API
- [ ] 默认按发布时间倒序

**阶段产出**：客户端可查询和浏览文章。

## Phase 3：稳定性与工程化（2~4 天）

- [ ] 增加结构化日志
- [ ] 增加 `healthz` / `readyz`
- [ ] 覆盖核心逻辑单测
- [ ] 配置 CI（`go test` + `go vet`）

**阶段产出**：具备基础可维护性与持续交付能力。

## Phase 4：增强能力（持续迭代）

- [ ] 多源抓取（RSS + HTML + 第三方 API）
- [ ] 标签与主题聚类
- [ ] 热门新闻评分
- [ ] 多语言支持

## Definition of Done（每阶段验收）

- [ ] 可运行（本地一条命令启动）
- [ ] 有测试（至少覆盖核心流程）
- [ ] 有文档（README/接口说明/变更说明）

