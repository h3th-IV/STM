## How to run the entire project

- **Local (single binary, HTTP + gRPC)**  
  ```bash
  go mod download
  go run ./cmd/api
  ```
  - HTTP API: `http://localhost:8080`  
  - gRPC (task notifications): `localhost:50051`

- **Demo gRPC client** (subscribe to task events for a user):  
  ```bash
  # Get JWT: POST /api/v1/auth/login, then:
  go run ./cmd/client -addr=localhost:50051 -user_id=1 -token=<access_token>
  ```
  Create/update/delete tasks via REST while the client is running to see events.

- **Docker**  
  ```bash
  docker build -t secure-task-go .
  docker run -p 8080:8080 -p 50051:50051 -e JWT_SECRET=your_secret secure-task-go
  ```

- **Regenerate proto (optional)**  
  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  export PATH="$(go env GOPATH)/bin:$PATH"
  protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/task.proto
  ```

## Deployment

Suitable for one-click deploy on:

- **Render**: Web Service, add `JWT_SECRET` env var
- **Fly.io**: `fly launch` then `fly secrets set JWT_SECRET=...`
- **Railway**: Connect repo, set env, deploy

For production, consider:

- Swap SQLite → PostgreSQL (change driver + connection string)
- Use Redis for refresh token storage
- Externalize secrets (Vault, AWS Secrets Manager)

## Future Improvements

- PostgreSQL support (easy driver swap)
- Redis for refresh token store
- CORS configuration
- OpenAPI/Swagger docs
- Request ID tracing