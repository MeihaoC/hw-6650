FROM --platform=linux/amd64 golang:1.25-alpine AS builder 

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o order-api ./cmd/order-api

# Final stage
FROM --platform=linux/amd64 alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/order-api .

EXPOSE 8080

CMD ["./order-api"]