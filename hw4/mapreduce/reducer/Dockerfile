FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY reducer.go .
COPY go.mod .
COPY go.sum .

RUN go mod download

# Build for Linux AMD64 to run on ECS Fargate
ENV GOOS=linux
ENV GOARCH=amd64
RUN go build -o reducer reducer.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/reducer .

EXPOSE 8082
CMD ["./reducer"]