FROM golang:1.19-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o bifrost-backups .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bifrost-backups .

CMD ["bifrost-backups"]