FROM golang:1.23.0-alpine AS builder

WORKDIR /app

COPY . .
RUN go mod download
RUN go build -o bifrost-backups .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bifrost-backups .

ENTRYPOINT ["/app/bifrost-backups"]