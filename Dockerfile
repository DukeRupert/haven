# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git curl nodejs npm

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@latest

# Install tailwindcss cli 
RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 && \
    chmod +x tailwindcss-linux-x64 && \
    mv tailwindcss-linux-x64 tailwindcss

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Generate templ files
RUN templ generate

# Run Tailwind CLI
RUN npx --yes tailwindcss -i ./web/assets/input.css -o ./web/assets/styles.css --minify

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Add necessary packages and create log directory
RUN apk add --no-cache ca-certificates tzdata && \
    mkdir -p /var/log/app && \
    chmod 755 /var/log/app

# Set environment variables for logging
ENV LOG_LEVEL=debug \
    ENVIRONMENT=production \
    TZ=UTC

# Copy binary and assets
COPY --from=builder /build/app .
COPY --from=builder /build/web/assets ./web/assets

# Create and switch to non-root user
RUN adduser -D appuser && \
    chown -R appuser:appuser /app /var/log/app
USER appuser

# Expose port
EXPOSE 8080

# Create volume for logs if needed
VOLUME ["/var/log/app"]

CMD ["./app"]
