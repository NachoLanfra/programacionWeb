-include .env
export

APP_NAME := programacion-web 

.PHONY: all generate build test clean docker

all: test

generate:
	@sqlc generate

docker:
	@docker compose up -d

build:
	@go build -o tmp/$(APP_NAME) .

test:
	@./test.sh

clean:
	@docker compose down -v --remove-orphans
	@rm -rf tmp
