FROM alpine:latest

SHELL ["/bin/ash", "-c"]
RUN apk add --no-cache postgresql-client

COPY bifrost-backups /bifrost-backups
ENTRYPOINT ["/bifrost-backups"]