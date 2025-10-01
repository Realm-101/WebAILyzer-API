# WebAIlyzer Lite API Makefile

.PHONY: build run test clean deps tidy fmt lint

# Build the application
build:
	go build -o bin/webailyzer-api ./cmd/webailyzer-api

# Run the application
run:
	go run ./cmd/webailyzer-api

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	go mod download

# Tidy dependencies
tidy:
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Run database migrations (requires migrate tool)
migrate-up:
	migrate -path migrations -database "postgres://webailyzer:@localhost:5432/webailyzer?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://webailyzer:@localhost:5432/webailyzer?sslmode=disable" down

# Create a new migration file
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations $$name

# Docker commands
docker-build:
	docker build -t webailyzer-lite-api .

docker-run:
	docker run -p 8080:8080 webailyzer-lite-api

# Development setup
dev-setup: deps tidy fmt

# Integration and benchmark tests
test-integration:
	@echo "Setting up test environment..."
	@docker-compose -f docker-compose.test.yml up -d postgres redis
	@sleep 5
	@echo "Running integration tests..."
	go test -v -tags=integration ./test/integration/... -timeout=30m
	@docker-compose -f docker-compose.test.yml down

test-benchmarks:
	@echo "Setting up benchmark environment..."
	@docker-compose -f docker-compose.test.yml up -d postgres redis
	@sleep 5
	@echo "Running benchmark tests..."
	go test -v -bench=. -benchmem ./test/benchmarks/... -timeout=60m
	@docker-compose -f docker-compose.test.yml down

test-e2e:
	@echo "Running end-to-end tests..."
	@docker-compose -f docker-compose.test.yml up -d
	@sleep 10
	@echo "Waiting for services to be ready..."
	@timeout 30 bash -c 'until curl -f http://localhost:8080/health; do sleep 2; done'
	go test -v -tags=e2e ./test/integration/... -timeout=45m
	@docker-compose -f docker-compose.test.yml down

test-performance:
	@echo "Running performance profiling..."
	@mkdir -p profiles
	go test -bench=BenchmarkAnalysisEndpoint -cpuprofile=profiles/cpu.prof ./test/benchmarks/...
	go test -bench=BenchmarkMemoryUsage -memprofile=profiles/mem.prof ./test/benchmarks/...
	@echo "Performance profiles saved to ./profiles/"

test-all: test test-integration test-benchmarks

# Test database setup
test-db-setup:
	@echo "Creating test databases..."
	@docker exec -i webailyzer_postgres createdb -U postgres webailyzer_test 2>/dev/null || true
	@docker exec -i webailyzer_postgres createdb -U postgres webailyzer_benchmark 2>/dev/null || true

test-db-cleanup:
	@echo "Cleaning up test databases..."
	@docker exec -i webailyzer_postgres dropdb -U postgres webailyzer_test 2>/dev/null || true
	@docker exec -i webailyzer_postgres dropdb -U postgres webailyzer_benchmark 2>/dev/null || true

# Full build and test
ci: deps tidy fmt lint test build

# CI with integration tests
ci-full: deps tidy fmt lint test test-integration build