#!/usr/bin/env bash
# Ciclo completo para probar el TP desde un clon limpio:
#   1. Tareas previas: borrar contenedores y volúmenes, generar código (templ + sqlc),
#      compilar, levantar PostgreSQL y esperar a que esté listo.
#   2. Ejecutar la aplicación hasta Ctrl+C.
#   3. Tareas posteriores: detener el servidor y borrar contenedores y volúmenes.
# Las tareas posteriores corren SIEMPRE (al terminar, al fallar o con Ctrl+C) gracias al trap.
set -euo pipefail

APP_BIN="./tmp/programacion-web"
SERVER_PID=""

cleanup() {
  local codigo=$?
  trap - EXIT INT TERM   # evita que cleanup se ejecute dos veces

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
# Ctrl+C o kill: salir con el código convencional (130 / 143); el trap EXIT limpia.
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
