# Build stage
FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the Go application (static binary)
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/web

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/main .

# Copy static assets and templates if needed at runtime
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates

# Expose the port your app listens on
EXPOSE 8080

# Run the application with default flags (can be overridden at runtime)
ENTRYPOINT ["./main"]
CMD ["-cache=false", "-production=false", "-cognito-user-pool-id=1234", "-cognito-client-id=1234"]
