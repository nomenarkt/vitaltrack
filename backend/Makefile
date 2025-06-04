build:
	go mod tidy
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./...

lint:
	@echo "==> Running linters..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./... ; \
	else \
		./bin/golangci-lint run ./... ; \
	fi
