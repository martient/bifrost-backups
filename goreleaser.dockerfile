FROM scratch
COPY bifrost-backups /bifrost-backups
ENTRYPOINT ["/bifrost-backups"]