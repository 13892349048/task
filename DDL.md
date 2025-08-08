```sql
-- users 表

CREATE TABLE `users` (
`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
`username` VARCHAR(64) NOT NULL,
`email` VARCHAR(255) DEFAULT NULL,
`password_hash` VARCHAR(128) NOT NULL,
`created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
`updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
PRIMARY KEY (`id`),
UNIQUE KEY `idx_users_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

  

-- tasks 表

CREATE TABLE `tasks` (
`id` BINARY(16) NOT NULL,
`user_id` BIGINT UNSIGNED NOT NULL,
`title` VARCHAR(255) NOT NULL,
`payload` JSON,
`priority` INT DEFAULT 0,
`status` VARCHAR(32) NOT NULL DEFAULT 'queued',
`result` JSON,
`retries` INT DEFAULT 0,
`created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
`updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
PRIMARY KEY (`id`),
KEY `idx_tasks_user_status` (`user_id`, `status`),
KEY `idx_tasks_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

  

-- task_logs 表（审计与重试记录）

CREATE TABLE `task_logs` (
`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
`task_id` BINARY(16) NOT NULL,
`attempt` INT NOT NULL DEFAULT 1,
`status` VARCHAR(32) NOT NULL,
`message` TEXT,
`created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
PRIMARY KEY (`id`),
KEY `idx_task_logs_task` (`task_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

  

-- idempotency_keys 表（保证幂等）

CREATE TABLE `idempotency_keys` (
`key` VARCHAR(128) NOT NULL,
`user_id` BIGINT UNSIGNED NULL,
`result_task_id` BINARY(16) NULL,
`created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
PRIMARY KEY (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

```

```yaml 

name: CI/CD

on:
  push:
    branches: [ "main", "staging" ]
  pull_request:
    branches: [ "main" ]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.20'
      - name: Build
        run: |
          go mod download
          go test ./... -v
          go build -o bin/app ./cmd/app
      - name: Run linters
        run: |
          go vet ./...
          golangci-lint run --timeout 5m
      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: app-bin
          path: bin/app

  docker:
    needs: build
    runs-on: ubuntu-latesty
    steps:
      - uses: actions/checkout@v4
      - name: Download artifact
        uses: actions/download-artifact@v4
        with:
          name: app-bin
      - name: Build Docker image
        run: |
          docker build -t ghcr.io/${{ github.repository_owner }}/task-dispatch:${{ github.sha }} .
      - name: Push image
        uses: docker/login-action@v2
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - run: |
          docker push ghcr.io/${{ github.repository_owner }}/task-dispatch:${{ github.sha }}

  deploy_staging:
    needs: docker
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/staging'
    steps:
      - uses: actions/checkout@v4
      - name: Deploy to staging server via SSH
        uses: appleboy/ssh-action@v0.1.7
        with:
          host: ${{ secrets.STAGING_HOST }}
          username: ${{ secrets.STAGING_USER }}
          key: ${{ secrets.STAGING_SSH_KEY }}
          script: |
            docker pull ghcr.io/${{ github.repository_owner }}/task-dispatch:${{ github.sha }}
            docker tag ghcr.io/${{ github.repository_owner }}/task-dispatch:${{ github.sha }} task-dispatch:latest
            docker-compose -f /opt/task-dispatch/docker-compose.staging.yml up -d --no-deps --build

  deploy_prod:
    needs: docker
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      - name: Manual approval
        uses: chrnorm/deployment-approval-action@v1
        with:
          reviewers: 'team-lead'
      - name: Deploy to prod via SSH
        uses: appleboy/ssh-action@v0.1.7
        with:
          host: ${{ secrets.PROD_HOST }}
          username: ${{ secrets.PROD_USER }}
          key: ${{ secrets.PROD_SSH_KEY }}
          script: |
            docker pull ghcr.io/${{ github.repository_owner }}/task-dispatch:${{ github.sha }}
            docker tag ghcr.io/${{ github.repository_owner }}/task-dispatch:${{ github.sha }} task-dispatch:latest
            docker-compose -f /opt/task-dispatch/docker-compose.prod.yml up -d --no-deps --build
```