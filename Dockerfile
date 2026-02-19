# Stage 1: Build
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api

# Stage 2: Runtime
FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/main .
RUN mkdir -p /app/data
ENV DB_PATH=/app/data/tasks.db
ENV PORT=8080
EXPOSE 8080
CMD ["./main"]
