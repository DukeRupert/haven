# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git curl nodejs npm

# Install templ and goose
RUN go install github.com/a-h/templ/cmd/templ@latest && \
    go install github.com/pressly/goose/v3/cmd/goose@latest

# Install tailwindcss cli 
RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64
RUN chmod +x tailwindcss-linux-x64
RUN mv tailwindcss-linux-x64 tailwindcss

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project
COPY . .

# Generate templ files
RUN templ generate

# Run Tailwind CLI
RUN npx --yes tailwindcss -i ./input.css -o ./styles.css --minify

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
COPY --from=builder /build/assets ./assets
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
