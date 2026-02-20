# STM — Secure Task Manager

Production-ready secure task management microservice in Go (Gin) with JWT authentication, RBAC, CRUD operations, and Dockerized DevSecOps pipeline. Demonstrates secure coding practices and containerization for cloud-native backends.

## Features

- **User authentication**: Registration, login, JWT access + refresh tokens with rotation
- **RBAC**: User vs admin roles (only admins can force-delete any task)
- **Task CRUD**: Create, read, update, delete personal tasks
- **Real-time task notifications**: gRPC server-streaming so clients can subscribe to task create/update/delete events (event-driven, low-latency).
- **Security**: bcrypt password hashing, rate limiting (auth routes), secure headers, input validation
- **OWASP mitigations**: Broken auth (JWT + bcrypt), injection (GORM), XSS, etc.
- **SQLite**: File-based database for easy Docker/local runs
- **Docker**: Multi-stage build, single container
- **CI/CD**: GitHub Actions with lint, test, build, Trivy vulnerability scan

## Tech Stack

| Component      | Package                                |
|----------------|----------------------------------------|
| Web framework  | `github.com/gin-gonic/gin` v1.10+     |
| ORM            | `gorm.io/gorm` + `gorm.io/driver/sqlite` |
| JWT            | `github.com/golang-jwt/jwt/v5`        |
| Password hash  | `golang.org/x/crypto/bcrypt`           |
| Rate limiting  | `github.com/ulule/limiter/v3`         |
| Validation     | `github.com/go-playground/validator/v10` |
| Logging        | `log/slog`                             |

## Quick Start

### Docker

```bash
docker build -t secure-task-go .
docker run -p 8080:8080 -p 50051:50051 -e JWT_SECRET=your_super_secret_key_change_me secure-task-go
```

Or with docker-compose:

```bash
docker-compose up --build
```

### Local

```bash
cp .env.example .env
# Edit .env and set JWT_SECRET
go mod download
go run ./cmd/api
```

- **HTTP API**: `http://localhost:8080`
- **gRPC (notifications)**: `localhost:50051`

## API Reference

Base URL: `http://localhost:8080/api/v1`

### Auth (Public)

| Method | Endpoint          | Description                          |
|--------|-------------------|--------------------------------------|
| POST   | `/auth/register`  | Register — `{email, password, username}` |
| POST   | `/auth/login`     | Login — `{email, password}`          |
| POST   | `/auth/refresh`   | Refresh tokens — `{refresh_token}`   |

### Protected (Bearer token required)

| Method | Endpoint      | Description                    |
|--------|---------------|--------------------------------|
| GET    | `/users/me`   | Current user profile           |
| GET    | `/tasks`      | List my tasks                  |
| POST   | `/tasks`      | Create task — `{title, description?, due_date?}` |
| GET    | `/tasks/:id`  | Get task                       |
| PUT    | `/tasks/:id`  | Update task                    |
| DELETE | `/tasks/:id`  | Delete task (owner or admin)   |

### Admin only

| Method | Endpoint           | Description             |
|--------|--------------------|-------------------------|
| DELETE | `/admin/tasks/:id` | Force delete any task   |

### Other

| Method | Endpoint | Description        |
|--------|----------|--------------------|
| GET    | `/health`| Health check       |

## gRPC — Real-time task notifications

Added real-time streaming notifications via gRPC to demonstrate event-driven, low-latency communication. When you create, update, or delete a task via REST, subscribed gRPC clients receive `TaskEvent` messages immediately.

- **Server**: Runs alongside Gin on **port 50051** (Gin stays on 8080).
- **Auth**: Same JWT as REST — send `authorization: Bearer <access_token>` in gRPC metadata.
- **RPC**: `NotificationService.SubscribeToTaskUpdates(SubscribeRequest) returns (stream TaskEvent)`.

### Run server (HTTP + gRPC)

```bash
go run ./cmd/api
# HTTP: :8080, gRPC: :50051
```

### Test with the demo client

1. Get a JWT (e.g. login via REST):
   ```bash
   curl -s -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"user@example.com","password":"password123"}' | jq -r .access_token
   ```
2. In one terminal, run the subscriber (use the same `user_id` as the token owner):
   ```bash
   go run ./cmd/client -addr=localhost:50051 -user_id=1 -token=<paste_access_token>
   ```
3. In another terminal, create a task via REST (same user):
   ```bash
   curl -X POST http://localhost:8080/api/v1/tasks \
     -H "Authorization: Bearer <access_token>" \
     -H "Content-Type: application/json" \
     -d '{"title":"New task","description":"Test"}'
   ```
   The client terminal should print a `CREATE` event.

### Test with grpcurl (optional)

```bash
# List services
grpcurl -plaintext localhost:50051 list

# Subscribe (requires -H "authorization: Bearer <token>" and valid user_id in request)
grpcurl -plaintext -d '{"user_id":"1"}' -H "authorization: Bearer <YOUR_JWT>" \
  localhost:50051 task.NotificationService/SubscribeToTaskUpdates
```

### Example cURL

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123","username":"johndoe"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'

# Create task (use access_token from login response)
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{"title":"My task","description":"Do something","due_date":"2025-12-31T23:59:59Z"}'

# List tasks
curl http://localhost:8080/api/v1/tasks -H "Authorization: Bearer <access_token>"
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Client                                │
└────────────────────────┬────────────────────────────────────┘
                         │ HTTP/REST
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                     Gin Router                               │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────────────┐ │
│  │ SecureHeaders│ │ RateLimit    │ │ Auth (JWT) + RBAC     │ │
│  └──────────────┘ └──────────────┘ └──────────────────────┘ │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│               Controllers (Auth, User, Task, Admin)           │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                    Services (Business Logic)                  │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│              Repositories (User, Task, RefreshToken)         │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                     SQLite (GORM)                            │
└─────────────────────────────────────────────────────────────┘
```

## Security Notes

Mitigates OWASP Top 10:

- **A01 Broken Auth**: JWT with secret, bcrypt password hashing, refresh token rotation
- **A02 Broken Auth (Cryptographic Failures)**: bcrypt for passwords
- **A03 Injection**: GORM parameterized queries, no raw SQL
- **A05 Security Misconfiguration**: Secure headers, CORS-ready
- **A07 Auth/Session Failures**: JWT expiry, refresh rotation

