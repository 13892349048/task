# Task Producer Service (MVP)

A Go service for high-concurrency task dispatch (producer-side API, MVP of Iteration 1).

## Quickstart (Windows PowerShell)

1. Set environment variables

```powershell
$env:TASK_ENVIRONMENT="dev"
$env:TASK_HTTP_PORT="8080"
$env:TASK_MYSQL_DSN="root:password@tcp(127.0.0.1:3306)/taskdb?charset=utf8mb4&parseTime=true&loc=Local"
$env:TASK_MYSQL_MAX_OPEN_CONNS="50"
$env:TASK_MYSQL_MAX_IDLE_CONNS="25"
$env:TASK_MYSQL_CONN_MAX_LIFETIME="30m"
$env:TASK_JWT_SECRET="dev-secret"
$env:TASK_JWT_ACCESS_TOKEN_TTL="1h"
```

2. Prepare database

- Create schema `taskdb` in MySQL and run `migrations/001_init.sql`

3. Run

```powershell
go run ./cmd/producer
```

4. Test API

- Register user
```powershell
curl -s -X POST http://127.0.0.1:8080/api/v1/users/register `
  -H "Content-Type: application/json" `
  -d '{"username":"u1","password":"p1","email":"u1@test.com"}'
```

- Login
```powershell
$token = (curl -s -X POST http://127.0.0.1:8080/api/v1/auth/login `
  -H "Content-Type: application/json" `
  -d '{"username":"u1","password":"p1"}') | ConvertFrom-Json
$env:JWT = $token.access_token
```

- Create task
```powershell
curl -s -X POST http://127.0.0.1:8080/api/v1/tasks `
  -H "Authorization: Bearer $env:JWT" `
  -H "Content-Type: application/json" `
  -d '{"title":"t1","payload":{"k":"v"},"priority":10}'
```

- Get task
```powershell
curl -s -H "Authorization: Bearer $env:JWT" http://127.0.0.1:8080/api/v1/tasks/<task_id>
```

- Health / Metrics
```powershell
curl -s http://127.0.0.1:8080/api/v1/health
curl -s http://127.0.0.1:8080/metrics
```

## Build Docker Image

```bash
docker build -t task-producer:dev .
```

## Notes
- This MVP writes tasks directly to MySQL. Kafka/Redis will be added in later iterations per design docs.
- Env var names follow `TASK_*` prefix and match `internal/config`. 