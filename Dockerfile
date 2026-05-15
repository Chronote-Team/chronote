FROM docker.io/library/golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/chronote ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/chronote-worker ./cmd/worker

FROM docker.io/library/alpine:3.22

RUN apk --no-cache add ca-certificates tzdata

RUN adduser -D -u 10001 chronote

WORKDIR /app

COPY --from=builder /out/chronote /app/chronote
COPY --from=builder /out/chronote-worker /app/chronote-worker

USER chronote

EXPOSE 8080

ENTRYPOINT ["/app/chronote"]
