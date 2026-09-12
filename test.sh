#!/usr/bin/env bash
set -euo pipefail

if [ -f .env ]; then
  set -a; 
  source .env; 
  set +a
fi

trap 'echo; echo "Limpiando"; make clean' EXIT

make clean         
make generate
make build
make docker

echo "Esperando a que la base de datos esté lista"

timeout=30; elapsed=0

until docker compose exec -T postgres pg_isready \
  -U "${POSTGRES_USER:-postgres}" -d "${POSTGRES_DB:-programacion_web}" \
  > /dev/null 2>&1
do
  sleep 1; elapsed=$((elapsed + 1))
  [ "$elapsed" -ge "$timeout" ] && { echo "Timeout esperando la base."; exit 1; }
done

go test ./... -v  
