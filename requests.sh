#!/usr/bin/env bash

set -uo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

PASSED=0
FAILED=0
STATUS=""
BODY=""
HEADERS=""

# Hace la petición con curl y deja el resultado en STATUS, BODY y HEADERS
request() {
  local method="$1" path="$2" data="${3:-}"
  local hdr_file out
  hdr_file="$(mktemp)"
  local args=(-s -S -X "$method" -D "$hdr_file" -o - -w '\n%{http_code}' "$BASE_URL$path")
  if [ -n "$data" ]; then
    args+=(-H 'Content-Type: application/json' -d "$data")
  fi
  out="$(curl "${args[@]}")" || { echo "  curl falló: $method $path"; rm -f "$hdr_file"; STATUS="000"; BODY=""; HEADERS=""; return; }
  STATUS="${out##*$'\n'}"
  BODY="${out%$'\n'*}"
  HEADERS="$(cat "$hdr_file")"
  rm -f "$hdr_file"
}

pass() { PASSED=$((PASSED + 1)); echo "  ok   $1"; }
fail() { FAILED=$((FAILED + 1)); echo "  FAIL $1"; echo "       status=$STATUS body=$BODY"; }

check() {
  local desc="$1" expected="$2" needle="${3:-}"
  if [ "$STATUS" != "$expected" ]; then
    fail "$desc (esperado $expected)"
  elif [ -n "$needle" ] && [[ "$BODY" != *"$needle"* ]]; then
    fail "$desc (el body no contiene: $needle)"
  else
    pass "$desc"
  fi
}

check_header() {
  if [[ "$HEADERS" == *"$2"* ]]; then pass "$1"; else FAILED=$((FAILED + 1)); echo "  FAIL $1"; echo "       headers=$HEADERS"; fi
}

json_int() { sed -n "s/.*\"$1\":\([0-9]*\).*/\1/p" <<<"$BODY" | head -n1; }

section() { echo; echo "== $1"; }

# Pruebas

section "Health y página estática"
request GET /health
check "GET /health" 200 '"status":"ok"'
if [ "$STATUS" != "200" ]; then
  echo
  echo "El servidor no responde en $BASE_URL. Está corriendo?"
  exit 1
fi
request GET /
check "GET / sirve la página estática" 200 'Plan De Estudios'

# USUARIOS 
section "Usuarios: POST /usuarios"
request POST /usuarios '{"nombre_apellido":"Ana Pérez","email":"ana@example.com","telefono":"2494111111"}'
check "crea un usuario" 201 '"email":"ana@example.com"'
USUARIO_ID="$(json_int id_usuario)"
check_header "devuelve el header Location" "Location: /usuarios/$USUARIO_ID"

request POST /usuarios '{"nombre_apellido":"   ","email":"x@example.com","telefono":"1"}'
check "rechaza nombre vacío" 400 'nombre_apellido'

request POST /usuarios '{"nombre_apellido":"Sin Email","email":"no-es-un-email","telefono":"1"}'
check "rechaza email inválido" 400 'email'

request POST /usuarios '{"nombre_apellido":"Sin Telefono","email":"st@example.com"}'
check "rechaza teléfono faltante" 400 'telefono'

request POST /usuarios '{"nombre_apellido":"Campo Extra","email":"ce@example.com","telefono":"1","admin":true}'
check "rechaza campos desconocidos" 400

request POST /usuarios '{esto no es json'
check "rechaza JSON mal formado" 400 'JSON'

request POST /usuarios ''
check "rechaza cuerpo vacío" 400

request POST /usuarios '{"nombre_apellido":"Otra Ana","email":"ana@example.com","telefono":"2"}'
check "rechaza email duplicado" 409 'email'

section "Usuarios: GET /usuarios y GET /usuarios/{id}"
request GET /usuarios
check "lista usuarios" 200 'ana@example.com'

request GET "/usuarios/$USUARIO_ID"
check "obtiene un usuario por id" 200 '"nombre_apellido":"Ana Pérez"'

request GET /usuarios/999999
check "id inexistente -> 404" 404 'no encontrado'

request GET /usuarios/abc
check "id no numérico -> 400" 400

request GET /usuarios/-5
check "id negativo -> 400" 400

section "Usuarios: PUT /usuarios/{id}"
request PUT "/usuarios/$USUARIO_ID" '{"nombre_apellido":"Ana María Pérez","email":"ana@example.com","telefono":"2494222222"}'
check "actualiza un usuario" 200 '"nombre_apellido":"Ana María Pérez"'

request GET "/usuarios/$USUARIO_ID"
check "el cambio quedó persistido" 200 '"telefono":"2494222222"'

request PUT "/usuarios/$USUARIO_ID" '{"nombre_apellido":"","email":"ana@example.com","telefono":"1"}'
check "PUT con datos inválidos -> 400" 400

request PUT /usuarios/999999 '{"nombre_apellido":"Nadie","email":"nadie@example.com","telefono":"1"}'
check "PUT de id inexistente -> 404" 404

# MATERIAS 
section "Materias: POST /materias"
request POST /materias '{"nombre":"Programación Web","anio":4,"cuatrimestre":1}'
check "crea una materia" 201 '"nombre":"Programación Web"'
MATERIA_ID="$(json_int id_materia)"
check_header "devuelve el header Location" "Location: /materias/$MATERIA_ID"

request POST /materias '{"nombre":"Algoritmos y Estructuras de Datos","anio":2,"cuatrimestre":1}'
check "crea otra materia" 201
MATERIA2_ID="$(json_int id_materia)"

request POST /materias '{"nombre":"","anio":1,"cuatrimestre":1}'
check "rechaza nombre vacío" 400 'nombre'

request POST /materias '{"nombre":"Año Cero","anio":0,"cuatrimestre":1}'
check "rechaza año fuera de rango" 400 'anio'

request POST /materias '{"nombre":"Cuatri Tres","anio":1,"cuatrimestre":3}'
check "rechaza cuatrimestre fuera de rango" 400 'cuatrimestre'

request POST /materias '{"nombre":"Año como texto","anio":"uno","cuatrimestre":1}'
check "rechaza tipos incorrectos" 400

section "Materias: GET /materias y GET /materias/{id}"
request GET /materias
check "lista materias" 200 'Programación Web'

request GET "/materias/$MATERIA_ID"
check "obtiene una materia por id" 200 '"anio":4'

request GET /materias/999999
check "id inexistente -> 404" 404

request GET /materias/abc
check "id no numérico -> 400" 400

section "Materias: PUT /materias/{id}"
request PUT "/materias/$MATERIA_ID" '{"nombre":"Programación Web (actualizada)","anio":4,"cuatrimestre":2}'
check "actualiza una materia" 200 '"cuatrimestre":2'

request PUT "/materias/$MATERIA_ID" '{"nombre":"Programación Web","anio":9,"cuatrimestre":1}'
check "PUT con datos inválidos -> 400" 400

request PUT /materias/999999 '{"nombre":"Fantasma","anio":1,"cuatrimestre":1}'
check "PUT de id inexistente -> 404" 404

# PROGRESO (usuario <-> materia)
section "Progreso: POST /usuarios/{id}/materias"
request POST "/usuarios/$USUARIO_ID/materias" "{\"materia_id\":$MATERIA_ID,\"estado\":\"cursando\"}"
check "registra una materia cursando (sin nota)" 201 '"estado":"cursando"'
check_header "devuelve el header Location" "Location: /usuarios/$USUARIO_ID/materias/$MATERIA_ID"

request POST "/usuarios/$USUARIO_ID/materias" "{\"materia_id\":$MATERIA_ID,\"estado\":\"aprobada\"}"
check "materia ya registrada -> 409" 409

request POST "/usuarios/$USUARIO_ID/materias" "{\"materia_id\":$MATERIA2_ID,\"estado\":\"quizas\"}"
check "estado inválido -> 400" 400 'estado'

request POST "/usuarios/$USUARIO_ID/materias" "{\"materia_id\":$MATERIA2_ID,\"estado\":\"aprobada\",\"nota\":11}"
check "nota fuera de rango -> 400" 400 'nota'

request POST "/usuarios/$USUARIO_ID/materias" '{"materia_id":999999,"estado":"cursando"}'
check "materia inexistente -> 422" 422

request POST /usuarios/999999/materias "{\"materia_id\":$MATERIA_ID,\"estado\":\"cursando\"}"
check "usuario inexistente -> 404" 404

request POST "/usuarios/$USUARIO_ID/materias" "{\"materia_id\":$MATERIA2_ID,\"estado\":\"aprobada\",\"nota\":8.5}"
check "registra una materia aprobada con nota" 201 '"nota":8.5'

section "Progreso: GET, PUT y DELETE"
request GET "/usuarios/$USUARIO_ID/materias"
check "lista el progreso del usuario" 200 'Programación Web'

request GET "/usuarios/$USUARIO_ID/materias/$MATERIA_ID"
check "obtiene el progreso en una materia" 200 '"estado":"cursando"'

request GET "/usuarios/$USUARIO_ID/materias/999999"
check "progreso inexistente -> 404" 404

request PUT "/usuarios/$USUARIO_ID/materias/$MATERIA_ID" '{"estado":"aprobada","nota":9}'
check "actualiza estado y nota" 200 '"estado":"aprobada"'

request GET "/usuarios/$USUARIO_ID/materias/$MATERIA_ID"
check "el cambio quedó persistido" 200 '"nota":9'

request PUT "/usuarios/$USUARIO_ID/materias/$MATERIA_ID" '{"estado":"inventado"}'
check "PUT con estado inválido -> 400" 400

request PUT "/usuarios/$USUARIO_ID/materias/999999" '{"estado":"cursando"}'
check "PUT de progreso inexistente -> 404" 404

request DELETE "/usuarios/$USUARIO_ID/materias/$MATERIA2_ID"
check "elimina un progreso" 204

request GET "/usuarios/$USUARIO_ID/materias/$MATERIA2_ID"
check "el progreso eliminado ya no existe" 404

request DELETE "/usuarios/$USUARIO_ID/materias/$MATERIA2_ID"
check "borrar de nuevo -> 404" 404

# CORRELATIVIDADES
section "Correlatividades: POST /materias/{id}/correlativas"
request POST "/materias/$MATERIA_ID/correlativas" "{\"materia_requerida_id\":$MATERIA2_ID}"
check "Programación Web requiere Algoritmos" 201 "\"id_materia_requerida\":$MATERIA2_ID"

request POST "/materias/$MATERIA_ID/correlativas" "{\"materia_requerida_id\":$MATERIA2_ID}"
check "correlatividad repetida -> 409" 409

request POST "/materias/$MATERIA2_ID/correlativas" "{\"materia_requerida_id\":$MATERIA_ID}"
check "correlatividad inversa (ciclo) -> 409" 409

request POST "/materias/$MATERIA_ID/correlativas" "{\"materia_requerida_id\":$MATERIA_ID}"
check "una materia no puede requerirse a sí misma -> 400" 400

request POST "/materias/$MATERIA_ID/correlativas" '{"materia_requerida_id":999999}'
check "materia requerida inexistente -> 422" 422

request POST /materias/999999/correlativas "{\"materia_requerida_id\":$MATERIA2_ID}"
check "materia inexistente -> 404" 404

request POST "/materias/$MATERIA_ID/correlativas" '{}'
check "falta materia_requerida_id -> 400" 400

section "Correlatividades: GET y DELETE"
request GET "/materias/$MATERIA_ID/correlativas"
check "lista las correlativas" 200 'Algoritmos y Estructuras de Datos'

request GET "/materias/$MATERIA2_ID/correlativas"
check "materia sin correlativas -> lista vacía" 200 '[]'

request GET /materias/999999/correlativas
check "GET de materia inexistente -> 404" 404

request DELETE "/materias/$MATERIA_ID/correlativas/$MATERIA2_ID"
check "elimina la correlatividad" 204

request DELETE "/materias/$MATERIA_ID/correlativas/$MATERIA2_ID"
check "borrar de nuevo -> 404" 404

request GET "/materias/$MATERIA_ID/correlativas"
check "ya no quedan correlativas" 200 '[]'

# BORRADO Y CASCADA
section "DELETE de entidades (y cascada)"
request DELETE "/materias/$MATERIA_ID"
check "elimina una materia" 204

request GET "/materias/$MATERIA_ID"
check "la materia eliminada ya no existe" 404

request GET "/usuarios/$USUARIO_ID/materias"
check "el progreso de esa materia se borró en cascada" 200 '[]'

request DELETE "/materias/$MATERIA_ID"
check "borrar una materia de nuevo -> 404" 404

request DELETE "/materias/$MATERIA2_ID"
check "elimina la otra materia" 204

request DELETE "/usuarios/$USUARIO_ID"
check "elimina un usuario" 204

request GET "/usuarios/$USUARIO_ID"
check "el usuario eliminado ya no existe" 404

request DELETE "/usuarios/$USUARIO_ID"
check "borrar un usuario de nuevo -> 404" 404

request DELETE /usuarios/abc
check "DELETE con id inválido -> 400" 400

section "Métodos no soportados"
request PATCH /usuarios
check "PATCH /usuarios -> 405" 405

echo
echo "Resultado: $PASSED verificaciones correctas, $FAILED fallidas."
[ "$FAILED" -eq 0 ]
