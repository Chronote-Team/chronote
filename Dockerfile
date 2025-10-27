FROM golang:1.25.3-alpine3.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root

COPY --from=builder /app/main .

RUN mkdir -p /app/config

EXPOSE 8080

ENV CONFIG_PATH=/app/config

CMD [ "./main" ]