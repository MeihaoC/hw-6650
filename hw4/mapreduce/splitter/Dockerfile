FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY splitter.go .
COPY go.mod .
COPY go.sum .

RUN go mod download
RUN go build -o splitter splitter.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/splitter .

EXPOSE 8080
CMD ["./splitter"]