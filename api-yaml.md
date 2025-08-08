```yaml
openapi: 3.0.3
info:
  title: High-Concurrency Task Dispatch API
  version: "1.0.0"
servers:
  - url: https://api.example.com
    description: Production
  - url: https://staging-api.example.com
    description: Staging
paths:
  /api/v1/auth/login:
    post:
      summary: User login
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                username:
                  type: string
                password:
                  type: string
              required: [username, password]
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  access_token:
                    type: string
                  token_type:
                    type: string
                  expires_in:
                    type: integer
        "401":
          description: Unauthorized

  /api/v1/users/register:
    post:
      summary: Register user
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                username:
                  type: string
                password:
                  type: string
                email:
                  type: string
              required: [username, password]
      responses:
        "201":
          description: Created
        "409":
          description: Conflict (username/email exists)

  /api/v1/tasks:
    post:
      summary: Create single task (async)
      parameters:
        - name: Idempotency-Key
          in: header
          schema:
            type: string
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                title:
                  type: string
                payload:
                  type: object
                priority:
                  type: integer
                  default: 0
                due_at:
                  type: string
                  format: date-time
              required: [title]
      responses:
        "202":
          description: Accepted
          content:
            application/json:
              schema:
                type: object
                properties:
                  task_id:
                    type: string
                  status:
                    type: string
        "503":
          description: Service Unavailable (e.g., Kafka write failed)

  /api/v1/tasks/batch:
    post:
      summary: Batch create tasks
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                tasks:
                  type: array
                  items:
                    type: object
                    properties:
                      title:
                        type: string
                      payload:
                        type: object
                      priority:
                        type: integer
      responses:
        "202":
          description: Accepted
          content:
            application/json:
              schema:
                type: object
                properties:
                  batch_id:
                    type: string

  /api/v1/tasks/{task_id}:
    get:
      summary: Get task status
      parameters:
        - name: task_id
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  task_id:
                    type: string
                  status:
                    type: string
                  result:
                    type: object
                  created_at:
                    type: string
                    format: date-time
                  updated_at:
                    type: string
                    format: date-time

  /api/v1/tasks/{task_id}/cancel:
    post:
      summary: Cancel a task
      parameters:
        - name: task_id
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: Cancelled
        "404":
          description: Not Found

  /api/v1/health:
    get:
      summary: Health check
      responses:
        "200":
          description: OK

  /metrics:
    get:
      summary: Prometheus metrics endpoint
      responses:
        "200":
          description: OK

components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
```