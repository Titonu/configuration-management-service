# Build stage
FROM golang:1.23.1-alpine AS builder
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled for SQLite support
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o config-service ./cmd/server

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite

WORKDIR /root/
COPY --from=builder /app/config-service .

# Create directory for SQLite database
RUN mkdir -p /root/data

# Expose the service port
EXPOSE 8080

# Run the service
CMD ["./config-service"]
