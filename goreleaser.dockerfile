FROM alpine:latest
COPY bifrost-backups /bifrost-backups
ENTRYPOINT ["/bifrost-backups"]