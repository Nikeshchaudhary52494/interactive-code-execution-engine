# Build stage
FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application
# CGO_ENABLED=0 creates a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -o execution-engine ./cmd/server

# Final stage
FROM alpine:latest

WORKDIR /app

# Install necessary runtime dependencies (if any)
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/execution-engine .

# Copy the frontend file
COPY index.html .

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./execution-engine"]
