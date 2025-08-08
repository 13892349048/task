#project-gpttask 

> 文档目的：把你提出的生产级技术考量（语言层面、缓存、MQ、数据库、业务容错等）整合为一份可执行的非功能性需求文档与升级架构图，作为后续接口设计、表结构和代码实现的蓝图。

---

## 1. 项目概述（回顾）

**项目**：即时任务派发系统（High‑Concurrency Task Dispatch System）
**目标**：支持互联网级高并发（目标示例：10k QPS 峰值、平均响应 <100ms）并保证系统稳定性、可观测性与可扩展性。

---

## 2. 可度量的非功能性需求（SLA / SLO / SLI）

**可用性 SLO**
- 99.9% 可用性（按月计算）

**性能 SLO / SLI**

- P95 响应时间 < 200ms（读取缓存）
- P95 写入 / 创建任务响应时间 < 500ms（采用异步派发）
- 支持单实例吞吐目标：2k QPS（通过水平扩容达到峰值）
    

**持久性 / 数据丢失**

- 关键业务（任务状态）持久化到 MySQL（事务保障）
- 异步消息使用 Kafka 持久化，消息一旦写入不会在未消费前丢失

**可观测性**

- 关键指标（QPS、平均延迟、error rate、GC pause、goroutine 数量、heap）暴露到 Prometheus
- Trace（分布式追踪）采样率：默认 5%（压力测试阶段可调高）

**可恢复性**

- 灾难恢复：DB 备份频率每日全量 + 每小时增量（RPO 1 小时）
- 自动重试策略与死信队列（DLQ）保证消息最终一致性

---

## 3. 语言层面（Go）设计细则

**目标**：控制 GC 和内存占用，降低延迟抖动，避免 goroutine 泄漏与过度并发。

### 3.1 GC 与内存优化

- 使用 `sync.Pool` 做短生命周期对象复用（如临时 `[]byte`、request buffer）。    
- 尽量**预分配切片容量**：`make([]T, 0, n)` 避免频繁扩容造成内存分配。
- 减少大对象短命分配：避免在热路径中频繁 new 结构体（use pool）。
- 避免使用接口抽象产生逃逸（interface 到堆）；若性能关键，使用具体类型。

### 3.2 减少内存逃逸

- 使用 `go vet -shadow` / `GODEBUG` / `pprof` 分析
- 返回值尽量为值类型（非指针），或确保局部变量不会逃逸

### 3.3 Goroutine 管理

- 设计**goroutine 池（worker pool）**限制最大并发数；不要为每个请求 spawn 无限制 goroutine。
- 使用 `context.Context` 做超时/取消控制，所有 goroutine 在退出路径上响应 cancel 信号。
- 对异步任务使用带缓冲的 channel 作速率缓冲，结合限流器（token bucket）进行节流。

### 3.4 阻塞与锁竞争分析

- 在关键路径开启 pprof 的 `block` profiling，分析锁等待。
- 使用分段锁 / sharding maps 来减少互斥范围；使用 `sync.RWMutex` 在读多写少场景。
- 避免长期持有锁（I/O / DB 操作不应在锁内）。

### 3.5 防内存泄漏

- 所有长期运行的 goroutine 需有生命周期管理（stop channel、context）。
- 关闭不需要的 channel；对 callback 注册要注销。
- 定期用 pprof/heap 做采样，设置报警：heap size 突增或 goroutine 数异常增长。

### 3.6 CPU 利用率控制

- 避免 busy‑wait；使用 `time.Ticker`/`select` 或带 sleep 的 retry。
- 结合 `runtime.GOMAXPROCS` 在容器场景根据 CPU 限制调优。

### 3.7 性能工具

- `pprof`（heap/cpu/block/goroutine）
- `go test -bench` 与基准测试
- `trace` 与 `runtime/metrics`（Go1.19+）

---

## 4. 缓存策略（Redis）

**目标**：高并发下使用多级缓存减少 DB 压力并保证数据正确性。

### 4.1 常见问题与方案

- **缓存穿透**：对不存在的 key 使用布隆过滤器或把空结果缓存（短 TTL）以拦截无效请求。
- **缓存击穿**：对热点 key 加互斥锁（singleflight）或用请求排队，预热热点缓存。
- **缓存雪崩**：给 key 设置随机化过期时间；分散热点 key 的过期时钟；多级缓存（本地 LRU + Redis）。

### 4.2 持久化配置（Redis）

- 推荐 **RDB + AOF 混合模式**：RDB 提供快速恢复快照，AOF 提供更精细的写日志。评估磁盘性能与 fsync 策略（`everysec` vs `always`）。
- 部署 Redis Cluster + 主从复制 + 哨兵/Cluster 模式，保证高可用。

### 4.3 缓存一致性策略

- **先写 DB 再删缓存（双删 + 延迟校验）**或使用 Redis 事务/锁确保顺序性。
- 对于强一致性场景，优先读 DB 并更新缓存或使用 CAS 式设计。

### 4.4 本地缓存

- 使用 LRU（`hashicorp/golang-lru`）或 `sync.Map` 做本地热点缓存，减少网络开销。
- 注意本地缓存失效带来的不一致性（适合近似一致性数据）。

---

## 5. 消息队列（Kafka 为首选）

**目标**：高吞吐、持久化、可回溯的异步派发机制。
### 5.1 选型理由

- **Kafka**：高吞吐、分区可并行消费、持久化日志、适合事件流与异步任务；适合互联网级流量。
- **RabbitMQ**：适合低延迟、复杂路由与事务场景，但在超大吞吐场景 Kafka 更合适。

### 5.2 持久化设计

- Kafka 默认落盘，配置 `min.insync.replicas` 严格度以保证写入持久化。
- 重要主题设置合适的 `retention` 与压缩（`log.retention.hours`、`cleanup.policy=compact`）策略。
- 消息队列确认机制 确保消息发送到消息队列
- 消费者确认机制 确保消费者完成消费 

### 5.3 顺序性与分区

- 顺序性保证：同一业务 key 映射到同一分区（partition by key）。
- 分区数规划：根据吞吐和并发消费者数量预估（分区数 >= consumer group size）。

### 5.4 消费模式与幂等

- 使用 **exactly‑once or at‑least‑once** 策略的权衡：多数业务采用 at‑least‑once + 幂等消费（消费端做去重）。
- 使用消费位点（offset）管理与事务化的 producer/consumer（Kafka Transactions）在需要时考虑。

### 5.5 失败处理

- 使用 **死信队列（DLQ）** 存放无法处理的消息；监控 DLQ 队列增长。
- 退避重试策略：消费失败采用指数退避并记录尝试次数。

### 5.6 监控与报警

- 监控指标：lag、produce/consume throughput, error count, ISR 状态
- 警报：lag 持续上升、broker offline、under‑replicated partitions

---

## 6. 数据库（MySQL）最佳实践

**目标**：保证高并发读写下的稳定性，避免连接泄漏与锁竞争。

### 6.1 连接池与空闲连接管理

- 配置 `SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`；`SetConnMaxLifetime` 推荐 < 应用容器最长停机时间，避免长期连接被 DB 回收导致错误。
- 避免过多空闲连接导致的资源占用，也避免频繁创建连接导致延迟。

### 6.2 事务与隔离级别

- 业务尽量使用短事务；长事务拆分或异步化。
- 选择合适隔离级别（默认 InnoDB 的可重复读），对读密集场景使用读已提交以减少锁等待。

### 6.3 索引设计与优化

- 使用覆盖索引避免回表（select 列只包含索引列）。
- 使用复合索引且顺序正确；避免在 where 子句使用函数导致索引失效。
- 定期分析慢查询（`EXPLAIN`、`pt‑query‑digest`）、建立索引和归档历史数据。

### 6.4 扩展策略

- 读写分离（主从复制）用于缓解读压
- 分库分表：对非常高的写吞吐使用水平拆分，使用一致性 hash 或业务路由（最后选择）因为涉及到分布式事务一致性问题，可能成本大大增加

### 6.5 热点与行级冲突

- 对热点行使用分散策略（时间分片、sharding）或批量累积操作减少频繁写。

---

## 7. 业务容错（熔断、降级、幂等）

### 7.1 超时与熔断

- 业务调用第三方或下游服务必须设置合理超时（context timeout），并使用熔断器（e.g. gobreaker）在失败率高时短路。    

### 7.2 降级策略

- 对非关键功能或统计数据，出现故障时返回默认值或缓存结果以保证主流程可用。

### 7.3 幂等性与重试

- 对写操作设计幂等 key（request id）和幂等处理逻辑；重试采用退避重试，避免洪峰时刻同时重试引发更大压力。

### 7.4 流量控制

- 对外部 API 加入限流（令牌桶）和熔断器；内部分布式限流可采用 Redis + Lua 实现全局计数器。    

---

## 8. 监控、日志与报警系统

**目标**：快速定位问题并自动化响应。
### 8.1 指标（Prometheus）

- 业务指标：QPS、error_rate、latency P50/P95/P99
- 系统指标：heap_alloc, gc_pause_ns, goroutines, cpu_seconds
- infra 指标：Redis hits/misses、Kafka lag、MySQL connections

### 8.2 日志（结构化）

- 使用 zap，输出 JSON 结构化日志：包含 trace_id, span_id, user_id, request_path, latency, error
- 日志集中化：ELK / Loki 用于查询与告警

### 8.3 分布式追踪

- 使用 Jaeger/OpenTelemetry 采集 trace；链路采样策略按流量调整

### 8.4 报警设计

- 报警分级：P0（系统不可用）、P1（功能严重受损）、P2（性能退化）
- 示例阈值：error_rate > 1% 且持续 5min → P1；P99 latency > 1s 持续 5min → P2

## 9. 部署、CI/CD 与可运维性

### 9.1 构建与镜像

- 使用多阶段 Dockerfile 生成瘦镜像；使用 image tags（commit SHA）保证可回溯。

### 9.2 部署策略

- 蓝绿 / 滚动 / 灰度发布（按流量分片），在生产先灰度 5% → 20% → 100%

### 9.3 配置管理

- 使用 Viper + 环境变量主导的配置（12‑factor app）；避免敏感信息出现在 repo，使用 Vault 或环境变量注入 Secrets。

### 9.4 回滚与回收

- 部署失败时自动回滚；保留历史镜像与备份。

## 10. 测试策略

- **单元测试**：核心算法、幂等、边界条件。
- **集成测试**：在 CI 环境部署依赖（MySQL、Redis、Kafka）做接口级测试。
- **性能测试**：使用 wrk / k6; 逐步提升并发到目标并分析瓶颈。
- **混沌测试**：模拟 Redis 故障、网络延迟、Kafka broker 下线，验证降级与恢复策略。

## 11. 运行手册（简要 Runbook 示例）

- **场景：API 响应率骤降**
    1. 检查 Prometheus 指标（QPS / latency / error_rate）
    2. 查看 pprof trace（block / cpu / heap）查看是否 GC 或锁竞争
    3. 查看 Redis/Kafka/MySQL 指标（connections、lag、slow queries）
    4. 若发现 Redis 宕机，触发降级：切换到 DB 读取并限制并发
- **场景：Kafka consumer lag 持续上升**
    1. 检查 consumer 错误日志与 DLQ
    2. 如果消费者出现 OOM/CPU 饱和，扩容消费者实例或缩减每实例线程数
    3. 若消息处理失败是业务异常，查看 DLQ 并手动补偿

## 12. 升级版架构图（ASCII 示意）

```
┌─────────── LoadBalancer ───────────┐
                     │               (Nginx)               │
                     └──────────────┬──────────────┬───────┘
                                    │              │
                     ┌──────────────▼──────────────┐┌───▼────────┐
                     │   App Instance (Docker)     ││ App (N)    │
                     │  - Gin                       ││ ...        │
                     │  - local LRU cache           ││            │
                     │  - worker pool / channels    ││            │
                     └──┬──────────┬────────────────┘└────────────┘
                        │          │
        ┌───────────────▼┐       ┌─▼──────────────┐
        │  Redis Cluster  │       │  Kafka Cluster │
        │  (cache + pub)  │       │ (topic partitions) │
        └───────┬─────────┘       └─────┬──────────┘
                │                       │
        ┌───────▼────────┐       ┌──────▼────────┐
        │   MySQL (P)    │◄──────│  Consumer(s)  │
        │   + replicas   │       │  (async workers)│
        └───────┬────────┘       └───────────────┘
                │
         ┌──────▼────────┐
         │  Backup / S3  │
         └───────────────┘

Monitoring & Observability:
- Prometheus ← metrics exporters
- Grafana dashboards
- Jaeger traces
- ELK / Loki logs

CI/CD: GitHub Actions / Jenkins -> build -> test -> image -> registry -> deploy
```

## 13. 迭代计划（结合非功能需求）

- **迭代 1（MVP）**：用户登录、任务创建（同步写 DB）、任务查询
- **迭代 2**：接入 Redis 缓存（查询走缓存），设置本地 LRU 层
- **迭代 3**：接入 Kafka，任务创建写入 Kafka（异步消费落库/派发）
- **迭代 4**：实现 worker pool、限流、熔断、降级策略
- **迭代 5**：全面监控（Prometheus/Grafana/Jaeger），压测与优化

