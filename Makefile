-include .env
export

APP_NAME := programacion-web 

.PHONY: all generate build test clean docker db-up db-down run

db-up:
	@docker compose up -d
	@echo "Esperando a que la base de datos esté lista..."
	@until docker exec postgres_db pg_isready -U $(POSTGRES_USER) >/dev/null 2>&1; do sleep 1; done
	@echo "Base de datos lista"

db-down:
	@docker compose down

run: db-up
	@go run .
	
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
