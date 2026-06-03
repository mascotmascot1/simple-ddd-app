# --- Build Stage ---
FROM golang:1.26.3 AS builder
WORKDIR /app

# Using separate COPY for go.mod/sum to leverage Docker layer caching
COPY go.sum go.mod ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 is used because modernc.org/sqlite is a pure Go implementation.
# This ensures a static binary and simplifies the final image environment.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server cmd/server/main.go 

# --- Final Image ---
FROM ubuntu:jammy
WORKDIR /app
RUN mkdir -p /app/data

# Copying only the binary to keep the final image size minimal and secure
COPY --from=builder /app/server .

# Note: All environment variables are provided at runtime via docker-compose or .env
# 
# Required variables for the Go app:
# HTTP_HOST, HTTP_PORT
# POSTGRES_HOST, POSTGRES_PORT, POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DATABASE, POSTGRES_SSLMODE
#
# Optional variables (with default values handled in Go code):
# HTTP_SHUTDOWN_TIMEOUT (default: 10) - in seconds
# ORDERS_HTTP_MAXUPLOADSIZE (default: 1000) - in KB

CMD ["./server"]
