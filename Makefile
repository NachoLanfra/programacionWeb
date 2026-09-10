.PHONY: all generate test clean docker

all: test

generate:
    @sqlc generate

docker:
    @docker compose up -d

test:
    @./test.sh

clean:
    @docker compose down -v --remove-orphans
