#!/usr/bin/env bash

set -euo pipefail
cd "$(dirname "$0")"

# Mismas credenciales que usa el Makefile para la base.
if [ -f .env ]; then
  set -a
  source .env
  set +a
fi

APP_BIN="./tmp/programacion-web"
SERVER_PID=""

cleanup() {
  local codigo=$?
  trap - EXIT INT TERM   # evita limpiar dos veces

  echo
  if [ -n "$SERVER_PID" ] && kill -0 "$SERVER_PID" 2>/dev/null; then
    echo "Deteniendo el servidor..."
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
  fi

  echo "Borrando contenedores y volúmenes..."
  make clean || true

  exit "$codigo"
}

trap cleanup EXIT
# Ctrl+C o kill: salir y limpiar.
trap 'exit 130' INT
trap 'exit 143' TERM

# 1. Tareas previas
make clean
make build
make db-up

# 2. Ejecutar la aplicación
echo "Aplicación en http://localhost:8080 (Ctrl+C para detenerla)"
"$APP_BIN" &
SERVER_PID=$!
wait "$SERVER_PID"
