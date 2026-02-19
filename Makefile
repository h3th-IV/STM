.PHONY: build run test lint docker

build:
	go build -o bin/api ./cmd/api

run:
	go run ./cmd/api

test:
	go test -v -race -coverprofile=coverage.out ./...

lint:
	golangci-lint run

docker:
	docker build -t secure-task-go .
	docker run -p 8080:8080 -e JWT_SECRET=your_super_secret_key_change_me secure-task-go
