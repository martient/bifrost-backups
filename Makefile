all: run

build:
	go build -o bin/bifrost-backups

run: build
	./bin/bifrost-backups