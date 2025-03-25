all: run

build:
	go build -o bin/bifrost-backups

run: build
	./bin/bifrost-backups

build-ubuntu:
	GOOS=linux GOARCH=amd64 go build -o bin/bifrost-backups-linux-amd64