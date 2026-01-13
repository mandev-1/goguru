FROM golang:1.21 AS builder

# Install build dependencies for CGO (required for sqlite3)
RUN apt-get update && apt-get install -y \
    gcc \
    libc6-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o camagru .

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    libsqlite3-0 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/camagru .

# Copy web files
COPY web ./web

# Create data directory structure
RUN mkdir -p data/uploads

EXPOSE 8080

CMD ["./camagru"]

