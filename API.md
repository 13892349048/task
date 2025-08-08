# Appendix A — 详细 API 设计（生产级）

> 下列接口包含路径、方法、请求/响应示例、常见错误码、幂等与限流设计建议。可直接转成 OpenAPI/Swagger。

## 认证模块

### POST /api/v1/auth/login

- 描述：用户登录，返回 JWT token
- 请求：
    

```
{
  "username": "string",
  "password": "string"
}
```

- 响应 200 OK：

```
{
  "access_token": "ey...",
  "token_type": "bearer",
  "expires_in": 3600
}
```

- 错误：401 Unauthorized（用户名/密码错误）
- 限流：每 IP 每分钟 60 次
    

### POST /api/v1/auth/refresh

- 描述：刷新 token（如使用 refresh token）
- 请求/响应略

---

## 用户模块

### POST /api/v1/users/register

- 描述：注册新用户
    
- 请求：
    

```
{
  "username":"string",
  "password":"string",
  "email":"string"
}
```

- 响应：201 Created
    
- 幂等设计：按 username/email 唯一键返回 409 Conflict
    

---

## 任务模块（核心）

### POST /api/v1/tasks

- 描述：创建单个任务（建议异步处理：写入 Kafka 并立即返回 202 Accepted）
- 请求头：`Authorization: Bearer <token>`；可选 `Idempotency-Key: <uuid>` 保证幂等
- 请求体：
    

```
{
  "title": "string",
  "payload": {"type":"object"},
  "priority": 10,
  "due_at": "2025-08-01T12:00:00Z"
}
```

- 响应：202 Accepted
    

```
{
  "task_id": "uuid",
  "status": "queued"
}
```

- 返回场景：创建时 producer 将消息写入 Kafka（确保写入成功后再返回）；若 Kafka 写入失败，返回 503 Service Unavailable 并触发本地重试队列
    
- 幂等：`Idempotency-Key` 存储于 Redis（ttl 24h），若重复请求返回相同 task_id
    
- 限流：每用户每秒 10 req；全局令牌桶限流
    

### POST /api/v1/tasks/batch

- 描述：批量创建任务，按批次拆分写入 Kafka，返回批次 ID
    
- 响应示例：202 Accepted + batch_id
    

### GET /api/v1/tasks/{task_id}

- 描述：查询任务状态（优先读 Redis 缓存，缓存未命中读 DB）
    
- 响应 200：
    

```
{
  "task_id":"uuid",
  "status":"running|queued|done|failed",
  "result": { },
  "created_at":"...",
  "updated_at":"..."
}
```

- 缓存：任务状态写入时同时更新 Redis，状态变更走消息/DB 后再删除或更新缓存
    

### GET /api/v1/tasks?user_id=&status=&limit=&offset=

- 列表接口，走缓存或分页读 DB
    

### POST /api/v1/tasks/{task_id}/cancel

- 取消任务：设置任务状态为 cancelled，通过消费者判断并停止执行；需要支持幂等
    

---

## 管理 / 运维 接口

### GET /api/v1/health

- 简单健康检查（返回 OK 与 downstream 状态概要）
    

### GET /api/v1/metrics

- 返回 Prometheus 指标暴露端点 `/metrics`