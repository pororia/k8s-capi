.PHONY: all build run test lint clean migrate docker-up docker-down web-dev web-build

BINARY_API=bin/api-server
BINARY_WORKER=bin/worker
BINARY_MIGRATOR=bin/migrator

# Build
all: build

build:
	go build -o $(BINARY_API) ./cmd/api-server
	go build -o $(BINARY_WORKER) ./cmd/worker
	go build -o $(BINARY_MIGRATOR) ./cmd/migrator

build-api:
	go build -o $(BINARY_API) ./cmd/api-server

build-worker:
	go build -o $(BINARY_WORKER) ./cmd/worker

build-migrator:
	go build -o $(BINARY_MIGRATOR) ./cmd/migrator

# Run
run:
	go run ./cmd/api-server

run-worker:
	go run ./cmd/worker

# Test
test:
	go test -v -race ./...

test-short:
	go test -v -short ./...

test-pkg:
	go test -v -run $(TEST) ./...

# Lint
lint:
	golangci-lint run ./...

# Format
fmt:
	gofmt -w .
	goimports -w .

# Vet
vet:
	go vet ./...

# Migration
migrate-up:
	go run ./cmd/migrator -action=up

migrate-down:
	go run ./cmd/migrator -action=down

migrate-status:
	go run ./cmd/migrator -action=status

# Docker (로컬 개발 환경)
docker-up:
	docker-compose -f deploy/docker-compose.yaml up -d

docker-down:
	docker-compose -f deploy/docker-compose.yaml down

docker-logs:
	docker-compose -f deploy/docker-compose.yaml logs -f

# Frontend
web-install:
	cd web && npm install

web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

web-lint:
	cd web && npm run lint

# Full stack (infra + backend + frontend)
stack-up:
	docker-compose -f deploy/docker-compose.full.yaml up -d

stack-down:
	docker-compose -f deploy/docker-compose.full.yaml down

# Clean
clean:
	rm -rf bin/

# Tidy
tidy:
	go mod tidy

# Generate mocks (mockery 사용 시)
gen-mocks:
	mockery --all --dir=internal/repository --output=internal/repository/mocks
