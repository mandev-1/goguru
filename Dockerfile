FROM golang:1.25.4-alpine AS builder

WORKDIR /app

# copy module files for Go cache
COPY go.mod go.sum ./
RUN go mod download

# copy full project
COPY . .

# build the server (explicit package path)
RUN go build -o /app/camagru ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/camagru .

# Copy web assets in the same relative layout the app expects
COPY --from=builder /app/web/static ./web/static
COPY --from=builder /app/web/templates ./web/templates

EXPOSE 8080

CMD ["./camagru"]