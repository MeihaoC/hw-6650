# Build stage 
FROM --platform=linux/amd64 golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY main.go .

# Build for linux/amd64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o product-api main.go

# Run stage
FROM --platform=linux/amd64 alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/product-api .

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["./product-api"]