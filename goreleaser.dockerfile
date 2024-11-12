FROM alpine:latest

RUN /bin/ash -c apk add --no-cache postgresql-client

COPY bifrost-backups /bifrost-backups
ENTRYPOINT ["/bifrost-backups"]