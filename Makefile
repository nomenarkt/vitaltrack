# Makefile for VitalTrack – Dev + CI Standards
# Usage: make up | make test | make check | make run | make docker-test

# ====================================
# 🐳 Docker Targets
# ====================================

up:
	docker compose up --build

down:
	docker compose down --remove-orphans --volumes

clean:
	@echo "🔪 Killing mock container (if running)..."
	-docker rm -f $(shell docker ps -aq --filter "name=airtable-mock") 2>/dev/null || true

rebuild: clean down up

docker-test:
	@echo "🧪 Running container smoke test..."
	curl --fail http://localhost:8787/health || (echo "❌ Service unavailable" && exit 1)

# ====================================
# 🧹 Backend Code Quality
# ====================================

lint:
	cd backend && golangci-lint run --timeout=2m

test:
	cd backend && go test ./... -v

coverage:
	cd backend && go test -coverprofile=coverage.out ./...

build:
	cd backend && go build -o bin/server ./cmd/server

check: lint test build

# ====================================
# 🔁 Developer Tools
# ====================================

run:
	cd backend && air

reset-db:
	@echo "TODO: implement database reset (via migrate or SQL)"

lint-version:
	@golangci-lint --version | grep "1.64.8" || (echo "❌ golangci-lint not at expected version" && exit 1)

# ====================================
# 🧪 CI/CD Utility Targets
# ====================================

ci:
	make check
	make coverage

.PHONY: up down clean rebuild docker-test lint test coverage build check run reset-db lint-version ci


#| Command             | Description                                 |
#| ------------------- | ------------------------------------------- |
#| `make docker-test`  | Curl healthcheck endpoint                   |
#| `make coverage`     | Generate `coverage.out` for CI Codecov      |
#| `make lint-version` | Ensure correct lint version is installed    |
#| `make ci`           | Full quality check + coverage for pipelines |
