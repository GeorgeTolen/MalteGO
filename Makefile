BINARY     := maltego-server
CMD        := ./cmd/server
PKG        := ./internal/...
PORT       ?= 8080

.PHONY: all build run test test-verbose test-cover lint fmt vet \
        docker-build docker-up docker-down docker-logs clean help

# ──────────────────────────────────────────────────────────────
# DEFAULT: build + test
# ──────────────────────────────────────────────────────────────
all: fmt vet test build

# ──────────────────────────────────────────────────────────────
# BUILD
# ──────────────────────────────────────────────────────────────
build:
	@echo ">>> Building $(BINARY)..."
	go build -ldflags="-s -w" -o $(BINARY) $(CMD)
	@echo ">>> Done: ./$(BINARY)"

# ──────────────────────────────────────────────────────────────
# RUN (local, без Docker)
# ──────────────────────────────────────────────────────────────
run:
	@echo ">>> Starting server on :$(PORT)..."
	go run $(CMD)

run-binary: build
	@echo ">>> Running ./$(BINARY) on :$(PORT)..."
	./$(BINARY)

# ──────────────────────────────────────────────────────────────
# TESTS
# ──────────────────────────────────────────────────────────────
test:
	@echo ">>> Running tests..."
	go test $(PKG) -count=1

test-verbose:
	@echo ">>> Running tests (verbose)..."
	go test $(PKG) -v -count=1

test-cover:
	@echo ">>> Running tests with coverage..."
	go test $(PKG) -count=1 -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out
	@echo ">>> HTML report: coverage.html"
	go tool cover -html=coverage.out -o coverage.html

test-race:
	@echo ">>> Running tests with race detector..."
	go test $(PKG) -race -count=1

# ──────────────────────────────────────────────────────────────
# CODE QUALITY
# ──────────────────────────────────────────────────────────────
fmt:
	@echo ">>> Formatting..."
	go fmt $(PKG) $(CMD)

vet:
	@echo ">>> Vetting..."
	go vet $(PKG) $(CMD)

lint:
	@echo ">>> Linting (requires golangci-lint)..."
	golangci-lint run $(PKG) $(CMD)

# ──────────────────────────────────────────────────────────────
# DOCKER
# ──────────────────────────────────────────────────────────────
docker-build:
	@echo ">>> Building Docker image..."
	docker-compose build

docker-up:
	@echo ">>> Starting service in Docker..."
	docker-compose up -d
	@echo ">>> Service running on http://localhost:$(PORT)"

docker-down:
	@echo ">>> Stopping Docker service..."
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-restart: docker-down docker-up

# Full cycle: build image, run, show logs
docker-run: docker-build docker-up docker-logs

# ──────────────────────────────────────────────────────────────
# CLEANUP
# ──────────────────────────────────────────────────────────────
clean:
	@echo ">>> Cleaning..."
	go clean
	rm -f $(BINARY) coverage.out coverage.html

# ──────────────────────────────────────────────────────────────
# HELP
# ──────────────────────────────────────────────────────────────
help:
	@echo ""
	@echo "MalteGO — GreyNoise Maltego Transform Server"
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "  all            fmt + vet + test + build (default)"
	@echo ""
	@echo "  Build:"
	@echo "  build          Compile binary ./$(BINARY)"
	@echo ""
	@echo "  Run:"
	@echo "  run            go run (hot, без компиляции)"
	@echo "  run-binary     build + запустить бинарник"
	@echo "  docker-run     build image + up + logs"
	@echo "  docker-up      Запустить в Docker (detached)"
	@echo "  docker-down    Остановить Docker-контейнер"
	@echo "  docker-logs    Показать логи контейнера"
	@echo "  docker-restart Перезапустить контейнер"
	@echo ""
	@echo "  Tests:"
	@echo "  test           go test ./internal/..."
	@echo "  test-verbose   с -v флагом"
	@echo "  test-cover     с отчётом покрытия (coverage.html)"
	@echo "  test-race      с детектором гонок"
	@echo ""
	@echo "  Quality:"
	@echo "  fmt            go fmt"
	@echo "  vet            go vet"
	@echo "  lint           golangci-lint (требует установки)"
	@echo ""
	@echo "  clean          Удалить бинарник и coverage-файлы"
	@echo ""
