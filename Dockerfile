# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS builder

# Install git (required for go mod) + curl for templ
RUN apk add --no-cache git curl

WORKDIR /app

# Only copy go mod/sum first to cache deps
COPY go.mod go.sum ./
RUN go mod download

# Install templ CLI once
RUN curl -fsSL https://raw.githubusercontent.com/a-h/templ/main/install.sh | sh

# Copy the rest of your source code *after* deps downloaded
COPY . .

# Only regenerate templ files if changed
RUN ./bin/templ generate

# Build binary
RUN go build -o app .

# -- Production stage --
FROM alpine:latest

RUN adduser -D appuser
USER appuser

COPY --from=builder /app/app /usr/local/bin/app

EXPOSE 8090
CMD ["app"]
