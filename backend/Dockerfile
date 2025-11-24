# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

ARG APP_VERSION=0.0.0

# Copy go mod files
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code
COPY . .

# Build the application with cache mounts and version injection
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X 'easi/backend/internal/infrastructure/api.Version=${APP_VERSION}'" \
    -o /app/api cmd/api/main.go

# Run stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/api .

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 8080

CMD ["./api"]
