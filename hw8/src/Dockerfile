# Build stage
FROM --platform=linux/amd64 golang:1.23-alpine AS builder 

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o shopping-cart-service .

# Runtime stage
FROM --platform=linux/amd64 alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/shopping-cart-service .

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["./shopping-cart-service"]