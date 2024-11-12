FROM alpine:latest

RUN apk add --no-cache postgresql-client

COPY bifrost-backups /bifrost-backups
ENTRYPOINT ["/bifrost-backups"]