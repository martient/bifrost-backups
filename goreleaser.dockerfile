FROM --platform=$BUILDPLATFORM alpine:latest

SHELL ["/bin/ash", "-c"]
RUN apk add --no-cache postgresql-client

ARG TARGETARCH
COPY bifrost-backups_linux_${TARGETARCH}/bifrost-backups /bifrost-backups
ENTRYPOINT ["/bifrost-backups"]