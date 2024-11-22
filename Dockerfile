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
RUN npx --yes tailwindcss -i ./input.css -o ./styles.css --minify

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /build/app .
COPY --from=builder /build/assets ./assets

RUN adduser -D appuser
USER appuser

EXPOSE 8080

CMD ["./app"]