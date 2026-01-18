# Production build (excludes Swagger documentation)
FROM golang:1.25.3-alpine3.22 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -tags prod -ldflags="-s -w" -o main .

# Runtime image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN apk --no-cache add tzdata
WORKDIR /root
COPY --from=builder /app/main .

RUN mkdir -p /app/config

EXPOSE 8080

ENV CONFIG_PATH=/app/config
ENV GIN_MODE=release

CMD ["./main"]
