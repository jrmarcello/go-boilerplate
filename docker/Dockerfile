# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Download dependencies first (cache layer)
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./cmd/api

# =============================================================================
# Production Stage - Appmax Standard
# =============================================================================
FROM debian:bullseye-slim@sha256:75e0b7a6158b4cc911d4be07d9f6b8a65254eb8c58df14023c3da5c462335593

RUN apt-get update && apt-get install -y git build-essential gcc libc6-dev pkg-config libsqlite3-dev tzdata ca-certificates && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# Configura o timezone para São Paulo
ENV TZ=America/Sao_Paulo
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Create unprivileged user
RUN mkdir -p /home/appuser && \
    echo "appuser:x:1000:1000::/home/appuser:/usr/sbin/nologin" >> /etc/passwd && \
    echo "appuser:x:1000:" >> /etc/group

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/main .

# Copy migrations
COPY --from=builder /build/internal/infrastructure/db/postgres/migration ./migrations

RUN chmod +x main && mkdir -p data && chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

CMD ["./main"]
