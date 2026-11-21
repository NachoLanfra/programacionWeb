-include .env
export

APP_NAME := programacion-web
POSTGRES_USER ?= postgres
POSTGRES_DB ?= programacion_web
# templ se instala en el bin de Go, que puede no estar en el PATH.
TEMPL := $(shell go env GOPATH)/bin/templ

.PHONY: all generate templ build test clean docker db-up db-down run run-app seed

# Levanta Postgres y espera (hasta 60 s) a que acepte conexiones por TCP.
# Se chequea por TCP (-h 127.0.0.1) porque mientras Postgres corre el esquema y
# la seed de la primera vez solo escucha por socket local: así la app no
# arranca antes de que la base esté cargada.
db-up:
	@docker compose up -d postgres
	@echo "Esperando a que la base de datos esté lista..."
	@for i in $$(seq 1 60); do \
		docker exec postgres_db pg_isready -h 127.0.0.1 -U $(POSTGRES_USER) >/dev/null 2>&1 && break; \
		[ $$i -eq 60 ] && { echo "Timeout esperando la base de datos"; exit 1; }; \
		sleep 1; \
	done
	@echo "Base de datos lista"

db-down:
	@docker compose down

seed: db-up
	@docker exec -i postgres_db psql -q -U $(POSTGRES_USER) -d $(POSTGRES_DB) < db/seed/seed.sql
	@echo "Plan de estudios cargado"

# Ciclo completo para probar el TP desde un clon limpio (ver run-app.sh):
# limpia, genera, compila, levanta la base, ejecuta la app y, al terminar o
# fallar, detiene el servidor y borra contenedores y volúmenes.
run-app:
	@./run-app.sh

# Uso diario: genera el código, levanta la base (sin borrar datos) y arranca el servidor.
run: generate db-up
	@go run .
	
all: test

# Genera el código Go de las queries (sqlc) y de las vistas (templ).
generate: templ
	@sqlc generate

templ:
	@$(TEMPL) generate

docker:
	@docker compose up -d postgres

build: generate
	@mkdir -p tmp
	@go build -o tmp/$(APP_NAME) .

test:
	@./test.sh

clean:
	@docker compose down -v --remove-orphans
	@rm -rf tmp
