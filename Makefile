.PHONY: all docker generate test

all: docker generate

docker:
	docker compose up -d

generate:
	sqlc generate
