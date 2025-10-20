# Build stage 
FROM --platform=linux/amd64 golang:1.25-alpine AS builder 

# Set working directory 
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application for linux/amd64 
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o product-search-service main.go

# Runtime stage 
FROM --platform=linux/amd64 alpine:latest

# Install ca-certificates for HTTPS requests (if needed)
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/product-search-service .

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["./product-search-service"]