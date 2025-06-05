# Start with a minimal Go image
FROM golang:1.21-alpine AS builder

# Enable go mod and templ
ENV CGO_ENABLED=0 \
    GO111MODULE=on \
    TEMPDIR=/tmp

# Install required tools and dependencies
RUN apk add --no-cache git bash curl

# Set the working directory
WORKDIR /app

# Cache go mod files first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Cache templ and templui CLI installation
RUN go install github.com/a-h/templ/cmd/templ@latest

# Copy everything else (source files, templates, etc.)
COPY . .

# Precompile templ files (cached if unchanged)
RUN templ generate

# Build the Go application
RUN go build -o app .

# --- Final image ---
FROM alpine:latest

# Set up non-root user (optional but recommended)
RUN adduser -D appuser
USER appuser

# Copy binary from builder
COPY --from=builder /app/app /usr/local/bin/app

# Copy static/assets if needed
# COPY --from=builder /app/static /app/templates /desired/path

# Expose and run
EXPOSE 8080
CMD ["/usr/local/bin/app"]
