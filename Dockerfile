FROM docker.io/library/golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/chronote ./cmd/api

FROM docker.io/library/alpine:3.22

RUN apk --no-cache add ca-certificates tzdata

RUN adduser -D -u 10001 chronote

WORKDIR /app

COPY --from=builder /out/chronote /app/chronote

USER chronote

EXPOSE 8080

ENTRYPOINT ["/app/chronote"]
