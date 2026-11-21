#!/usr/bin/env bash

set -euo pipefail
cd "$(dirname "$0")"

if [ -f .env ]; then
  set -a
  source .env
  set +a
fi

APP_BIN="./tmp/programacion-web"
LOG_FILE="$(mktemp)"
SERVER_PID=""
export BASE_URL="http://localhost:8080"

cleanup() {
  local codigo=$?
  trap - EXIT INT TERM

  if [ -n "$SERVER_PID" ] && kill -0 "$SERVER_PID" 2>/dev/null; then
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
  fi

  if [ "$codigo" -ne 0 ]; then
    echo
    echo "Falló (código $codigo). Últimos logs del servidor:"
    tail -n 30 "$LOG_FILE" 2>/dev/null || true
  fi
  rm -f "$LOG_FILE"

  echo
  echo "== Limpiando contenedores y volúmenes"
  make clean || true

  exit "$codigo"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

echo "== Limpiando restos de corridas anteriores"
make clean

echo "== Generando código y compilando"
make build

echo "== Levantando la base de datos"
make db-up

echo "== Levantando la aplicación"
"$APP_BIN" >"$LOG_FILE" 2>&1 &
SERVER_PID=$!

echo "== Esperando a que responda en ${BASE_URL}/health"
timeout=30; elapsed=0
until curl -fs "${BASE_URL}/health" >/dev/null 2>&1; do
  kill -0 "$SERVER_PID" 2>/dev/null || { echo "El servidor terminó antes de responder."; exit 1; }
  sleep 1; elapsed=$((elapsed + 1))
  [ "$elapsed" -ge "$timeout" ] && { echo "Timeout esperando la aplicación."; exit 1; }
done

echo "== Probando el recorrido HTML"
./requests-html.sh
