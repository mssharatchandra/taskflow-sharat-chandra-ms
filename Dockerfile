# Stage 1: Build the Go binary
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy dependency files first for better Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build a statically linked binary
RUN CGO_ENABLED=0 go build -o /taskflow ./cmd/server

# Stage 2: Minimal runtime image
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates curl postgresql-client

# Install golang-migrate for running database migrations
ARG TARGETARCH
RUN curl -L "https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-${TARGETARCH}.tar.gz" | tar xz \
    && mv migrate /usr/local/bin/migrate

COPY --from=builder /taskflow /usr/local/bin/taskflow
COPY migrations/ /migrations/
COPY scripts/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/entrypoint.sh"]
