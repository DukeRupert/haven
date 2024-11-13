# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git

# Install templ and goose
RUN go install github.com/a-h/templ/cmd/templ@latest && \
    go install github.com/pressly/goose/v3/cmd/goose@latest

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project
COPY . .

# Generate templ files
RUN templ generate

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/main.go

# Final stage
FROM golang:1.23-alpine

WORKDIR /app

# Install runtime dependencies and postgresql-client for goose
RUN apk add --no-cache ca-certificates postgresql-client

# Copy the binary and migrations
COPY --from=builder /build/app .
COPY --from=builder /build/migrations ./migrations
COPY --from=builder /go/bin/goose /usr/local/bin/goose

# Copy the startup script
COPY docker-entrypoint.sh .
RUN chmod +x docker-entrypoint.sh

# Create a non-root user but allow it to run goose
RUN adduser -D appuser
USER appuser

# Expose port
EXPOSE 8080

# Use the entrypoint script
ENTRYPOINT ["./docker-entrypoint.sh"]
