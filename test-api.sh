#!/usr/bin/env bash

set -euo pipefail
cd "$(dirname "$0")"

if [ -f .env ]; then
  set -a
  source .env
  set +a
fi

APP_PORT="${APP_PORT:-8080}"
export BASE_URL="http://localhost:${APP_PORT}"

cleanup() {
  status=$?
  if [ "$status" -ne 0 ]; then
    echo
    echo "Falló (código $status). Últimos logs de la app:"
    docker compose logs --tail=30 app 2>/dev/null || true
  fi
  echo
  echo "Limpiando contenedores y volúmenes"
  make clean
  exit "$status"
}
trap cleanup EXIT

echo "== Limpiando restos de corridas anteriores"
make clean

echo "== Generando código de acceso a datos (sqlc)"
make generate

echo "== Compilando"
make build

echo "== Levantando la base de datos"
docker compose up -d postgres

echo "== Esperando a que PostgreSQL esté listo"
timeout=60; elapsed=0
until docker compose exec -T postgres pg_isready \
  -U "${POSTGRES_USER:-postgres}" -d "${POSTGRES_DB:-programacion_web}" >/dev/null 2>&1
do
  sleep 1; elapsed=$((elapsed + 1))
  [ "$elapsed" -ge "$timeout" ] && { echo "Timeout esperando la base."; exit 1; }
done

echo "== Construyendo y levantando la aplicación"
docker compose up -d --build app

echo "== Esperando a que la API responda en ${BASE_URL}/health"
timeout=60; elapsed=0
until curl -fs "${BASE_URL}/health" >/dev/null 2>&1; do
  sleep 1; elapsed=$((elapsed + 1))
  [ "$elapsed" -ge "$timeout" ] && { echo "Timeout esperando la API."; exit 1; }
done

echo "== Probando la API"
./requests.sh
