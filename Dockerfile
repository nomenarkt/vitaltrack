FROM golang:1.24.3

# Set root working directory
WORKDIR /app

# Copy entire repo
COPY . .

# Change working directory to backend where go.mod lives
WORKDIR /app/backend

# Download dependencies
RUN go mod download

# Optional Dev Tools
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.1/migrate.linux-amd64.tar.gz \
  | tar xvz && mv migrate /usr/local/bin/migrate

# Build the binary
RUN go build -o server ./cmd/server

# Install golangci-lint
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install air (hot reload)
RUN curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b /usr/local/bin


CMD ["./server"]

