阶段总览（建议顺序）
阶段 0：项目对齐与环境准备（0.5 天）
阶段 1：MVP 同步写库的生产者 API（1–2 天）
阶段 2：配置与日志（0.5–1 天）
阶段 3：缓存层（Redis + 本地 LRU）（1–2 天）
阶段 4：接入 Kafka（异步生产者与幂等）（2–3 天）
阶段 5：可观测性（Prometheus/Tracing/结构化日志）（1–2 天）
阶段 6：韧性与限流（熔断/重试/令牌桶）（1–2 天）
阶段 7：CI/CD 与容器化（0.5–1 天）
阶段 8：压测与优化（1–2 天，穿插进行）
阶段 0：项目对齐与环境准备
学习目标
理解业务与非功能指标：延迟、吞吐、可靠性。
明确 MVP 边界（先不接入 Kafka/Redis）。
参考与阅读
High‑Concurrency Task Dispatch — 非功能性需求与升级架构.md（SLO/NFR/演进路线）
API.md、api-yaml.md（接口契约）
DDL.md（表结构）
产出
一页纸笔记：范围、关键指标、MVP 功能列表与不做的项。
验收标准
能口述系统目标、约束与第一阶段范围。
阶段 1：MVP 同步写库的生产者 API
学习目标
分层设计：Handler → Service → Repository → DB Model。
UUID BINARY(16) 存储与编码转换；JSON 字段的存取策略。
错误分类与 HTTP 映射（400/401/404/500/503）。
对照代码阅读
路由与中间件：internal/server/router.go
处理器：internal/handler/{auth_handler.go,task_handler.go,health_handler.go}
服务：internal/service/{auth_service.go,task_service.go}
仓库：internal/repository/{user_repo.go,task_repo.go,db.go}
模型：internal/model/{user.go,task.go}（BINARY(16) → UUID 字符串）
设计输出
时序图（请求到达→鉴权→服务→仓库→MySQL→响应）。
错误映射表与返回体规范（对照 API.md）。
验收标准
能清晰解释每一层的职责边界与返回值约定。
用样例请求推演一次创建任务和查询任务的完整链路。
阶段 2：配置与日志
学习目标
12-factor 配置：环境变量优先，默认值策略。
zap 日志级别、字段化、生产/开发编码器差异。
对照代码阅读
配置：internal/config/config.go（Viper 的键名与默认值）
日志：pkg/logger/logger.go + cmd/producer/main.go（初始化与使用）
设计输出
配置清单（必填/可选/默认）与上线前检查表。
日志字段规范：必须包含的字段与采样策略。
验收标准
能说明如何在不同环境切换配置与日志格式；能指出关键日志点位。
阶段 3：缓存层（Redis + 本地 LRU）
学习目标
缓存穿透/击穿/雪崩的应对策略；双删与延迟校验。
任务状态一致性与过期策略。
计划实现（不立刻写）：查询接口走缓存，写入时 DB→删/更新缓存。
设计输出
读写流程图；键空间设计；TTL 与随机抖动策略；失效时序。
验收标准
明确何时读缓存、何时落库回源、何时删/更新缓存。
阶段 4：接入 Kafka（异步生产者与幂等）
学习目标
生产者 ACK、min.insync.replicas、分区键选择与有序性。
幂等键（Idempotency-Key）与结果复用返回。
失败重试与 DLQ（策略与监控）。
计划实现（不立刻写）：
创建任务：写 Kafka 成功才返回 202；失败返回 503 并本地限次重试。
idempotency_keys 表或 Redis 记录请求键→task_id。
设计输出
Topic/分区/键策略；错误与重试矩阵；DLQ 处置 Runbook。
验收标准
能解释消息不丢、至少一次、幂等落库的保障点。
阶段 5：可观测性
学习目标
Prom 指标分层：系统、资源、业务；RED/USE 指标。
Trace 基本概念与采样；日志关联 TraceID。
计划实现（不立刻写）：
GET /metrics 暴露：QPS、错误率、延迟直方图、DB/缓存/Kafka 指标若干。
结构化日志包含 trace_id/user_id/path/latency/error。
设计输出
指标清单与告警阈值；Trace 采样方案。
验收标准
能根据指标面板定位常见瓶颈（DB 慢、Kafka 卡、缓存穿透）。
阶段 6：韧性与限流
学习目标
熔断（gobreaker）与退避重试；令牌桶限流（用户级/全局）。
上游/下游的超时与取消（context 传播）。
计划实现（不立刻写）：登录与任务创建的限流与熔断点位。
设计输出
熔断参数、恢复策略与灰度开关；限流配额策略。
验收标准
能对“下游不稳定”给出稳定降级方案。
阶段 7：CI/CD 与容器化
学习目标
GitHub Actions 基础、分阶段流水线与工件管理。
多阶段 Dockerfile、镜像瘦身、可观测性/安全基线。
对照文件
Dockerfile、.github/workflows/ci.yml
设计输出
分支策略、构建/测试/镜像/部署步骤与审批点。
验收标准
说明一次从提交到镜像推送的全流程。
阶段 8：压测与优化
学习目标
指标驱动优化：P95/P99 延迟、GC、goroutines、锁等。
分析工具：pprof、trace、runtime/metrics。
设计输出
压测场景与目标；优化前后对比与结论。
验收标准
能基于数据给出优化策略与权衡，避免过早优化。
学习与实践节奏建议
每阶段输出三件事：
设计/思考产出（简短笔记/图）
验收清单（自测问题 + 可观测项）
代码对照点位（明确要读的文件与关注点）
代码动手前，先过一遍“设计输出”和“验收标准”。