FROM ubuntu:latest

SHELL ["/bin/sh", "-c"]
RUN apt update -y && apt upgrade -y && apt install postgresql-client -y

COPY bifrost-backups /bifrost-backups
ENTRYPOINT ["/bifrost-backups"]