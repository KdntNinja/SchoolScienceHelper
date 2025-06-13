# Build-Stage
FROM golang:1.24-alpine AS build
WORKDIR /app

# Copy go mod/sum and download dependencies (cache layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Install templ (cacheable layer)
RUN --mount=type=cache,target=/go/pkg/mod go install github.com/a-h/templ/cmd/templ@latest

# Generate templ files
RUN templ generate

# Install build dependencies
RUN apk add gcc musl-dev

# Build the application (new entry point)
RUN CGO_ENABLED=1 GOOS=linux go build -o main ./cmd/server/main.go

# Deploy-Stage
FROM alpine:latest
WORKDIR /app

# Update package index and install ca-certificates
RUN apk update && apk upgrade && apk add ca-certificates

# Copy the binary from the build stage
COPY --from=build /app/main .
# Copy assets (migrations, static files) into the runtime image
COPY assets/ assets/

# Expose the port your application runs on
EXPOSE 8090

# Command to run the application
CMD ["./main"]