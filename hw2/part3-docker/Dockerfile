# Build the application from source
FROM golang:1.25-alpine AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /main

# Deploy the application binary into a lean image
FROM alpine:latest AS build-release-stage

RUN apk --no-cache add ca-certificates

WORKDIR /

COPY --from=build-stage /main /main

EXPOSE 8080

USER nobody:nobody

ENTRYPOINT ["/main"]