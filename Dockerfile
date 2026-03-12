# Production build
FROM golang:1.25.3-alpine3.22 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/main .

# Runtime image
FROM alpine:3.22
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/main ./main

EXPOSE 8080

ENV GIN_MODE=release

CMD ["./main"]
