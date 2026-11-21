#!/usr/bin/env bash

set -uo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

PASSED=0
FAILED=0
STATUS=""
BODY=""
HEADERS=""
LOCATION=""

# request METODO RUTA [campo=valor ...]
# Deja el resultado en STATUS, BODY, HEADERS y LOCATION.
request() {
  local method="$1" path="$2"
  shift 2
  local hdr_file out
  hdr_file="$(mktemp)"
  local args=(-s -S -X "$method" -D "$hdr_file" -o - -w '\n%{http_code}' "$BASE_URL$path")
  local campo
  for campo in "$@"; do
    args+=(--data-urlencode "$campo")
  done
  out="$(curl "${args[@]}")" || { echo "  curl falló: $method $path"; rm -f "$hdr_file"; STATUS="000"; BODY=""; HEADERS=""; LOCATION=""; return; }
  STATUS="${out##*$'\n'}"
  BODY="${out%$'\n'*}"
  HEADERS="$(cat "$hdr_file")"
  LOCATION="$(sed -n 's/^[Ll]ocation: //p' "$hdr_file" | tr -d '\r')"
  rm -f "$hdr_file"
}

# hace el POST del formulario de la fila de esa materia.
enviar() {
  request POST /cursadas "materia_id=$(id_materia "$1")" "estado=$2" "nota=$3"
}

# hace el GET al que redirigió el último POST (la "G" de PRG).
seguir() {
  request GET "$LOCATION"
}

pass() { PASSED=$((PASSED + 1)); echo "  ok   $1"; }
fail() { FAILED=$((FAILED + 1)); echo "  FAIL $1"; echo "       status=$STATUS location=$LOCATION"; }

# verifica el código HTTP y, opcionalmente, que el body contenga TEXTO.
check() {
  local desc="$1" expected="$2" needle="${3:-}"
  if [ "$STATUS" != "$expected" ]; then
    fail "$desc (esperado $expected)"
  elif [ -n "$needle" ] && [[ "$BODY" != *"$needle"* ]]; then
    fail "$desc (el HTML no contiene: $needle)"
  else
    pass "$desc"
  fi
}

# verifica que el body NO contenga TEXTO.
check_not() {
  if [[ "$BODY" == *"$2"* ]]; then fail "$1 (el HTML contiene: $2)"; else pass "$1"; fi
}

# verifica un 303 See Other cuyo Location empieza con DESTINO.
check_redirect() {
  local desc="$1" destino="$2"
  if [ "$STATUS" != "303" ]; then
    fail "$desc (esperado 303)"
  elif [[ "$LOCATION" != "$destino"* ]]; then
    fail "$desc (Location no empieza con $destino)"
  else
    pass "$desc"
  fi
}

# después de un POST inválido, verifica que redirija con error y que la página muestre el aviso para MATERIA con TEXTO.
check_aviso() {
  local desc="$1" materia="$2" texto="$3"
  check_redirect "$desc: redirige con error" "/?"
  if [[ "$LOCATION" != *"error="* ]]; then
    fail "$desc: la redirección no trae el error"
    return
  fi
  seguir
  check "$desc: muestra el aviso" 200 'class="error"'
  if [ -n "$materia" ]; then
    check "$desc: el aviso nombra la materia" 200 "No se pudo guardar $materia"
  fi
  check "$desc: el aviso explica el error" 200 "$texto"
}

# imprime el HTML de la fila de la tabla de esa materia (de la última página pedida).
fila() {
  sed 's#</tr>#</tr>\n#g' <<<"$BODY" | grep -F "<td>$1</td>" | head -n1
}

# pide / y verifica el estado (y la nota) de una fila.
check_fila() {
  local desc="$1" materia="$2" estado="$3" nota="${4:-}"
  request GET /
  local html
  html="$(fila "$materia")"
  if [ -z "$html" ]; then
    fail "$desc (no se encontró la fila de $materia)"
  elif [[ "$html" != *"value=\"$estado\" selected"* ]]; then
    fail "$desc (la fila no tiene $estado seleccionado)"
  elif [ -n "$nota" ] && [[ "$html" != *"name=\"nota\""*"value=\"$nota\""* ]]; then
    fail "$desc (la fila no tiene nota $nota)"
  elif [ -z "$nota" ] && [[ "$html" != *"name=\"nota\""*"value=\"\""* ]]; then
    fail "$desc (la fila tiene nota y no debería)"
  else
    pass "$desc"
  fi
}

# busca el id de la materia en la base (los ids no son fijos).
id_materia() {
  docker exec postgres_db psql -tA -U "${POSTGRES_USER:-postgres}" -d "${POSTGRES_DB:-programacion_web}" \
    -c "SELECT id_materia FROM materia WHERE nombre = '$1'"
}

section() { echo; echo "== $1"; }

# --- Pruebas ---------------------------------------------------------------

section "Health"
request GET /health
check "GET /health" 200 "ok"
if [ "$STATUS" != "200" ]; then
  echo
  echo "El servidor no responde en $BASE_URL. ¿Está corriendo?"
  exit 1
fi

section "Página principal: GET /"
request GET /
check "responde 200" 200
check "es la página renderizada con el layout" 200 "<title>Plan De Estudios</title>"
check "carga Pico.css" 200 "pico.min.css"
check "muestra el alumno" 200 "Alumno de Prueba"
FILAS="$(grep -o '<td><form method="POST" action="/cursadas">' <<<"$BODY" | wc -l)"
if [ "$FILAS" -eq 38 ]; then pass "la tabla tiene las 38 materias del plan"; else fail "la tabla tiene $FILAS filas (esperado 38)"; fi
check_not "no hay aviso de error" 'class="error"'

section "Datos de la seed"
check_fila "Introducción a la Programación I finalizada con 8" "Introducción a la Programación I" finalizada 8
check_fila "Cálculo I regular con 5" "Cálculo I" regular 5
check_fila "Programación Orientada a Objetos cursando, sin nota" "Programación Orientada a Objetos" cursando
check_fila "Cálculo II sin cursar, sin nota" "Cálculo II" sin_cursar

section "Guardar una cursada válida (Post/Redirect/Get)"
enviar "Cálculo II" cursando ""
check_redirect "POST Cálculo II -> cursando: 303 a /" "/"
if [ "$LOCATION" = "/" ]; then pass "la redirección no trae error"; else fail "la redirección trae algo más que /"; fi
check_fila "la página muestra Cálculo II cursando" "Cálculo II" cursando

enviar "Cálculo II" regular 7
check_redirect "POST Cálculo II -> regular con 7: 303 a /" "/"
check_fila "la página muestra Cálculo II regular con 7" "Cálculo II" regular 7

enviar "Cálculo II" regular 7.5
check_redirect "POST con nota decimal (7.5): 303 a /" "/"
check_fila "la página muestra la nota 7.5" "Cálculo II" regular 7.5

section "Validaciones de estado y nota"
enviar "Base de Datos II" sin_cursar 7
check_aviso "sin cursar con nota" "Base de Datos II" "no puede tener nota"

enviar "Inglés" cursando 6
check_aviso "cursando con nota" "Inglés" "no puede tener nota"

enviar "Cálculo II" regular ""
check_aviso "regular sin nota" "Cálculo II" "tenés que cargar la nota"

enviar "Cálculo II" finalizada ""
check_aviso "finalizada sin nota" "Cálculo II" "tenés que cargar la nota"

enviar "Cálculo II" regular 3
check_aviso "nota menor a 4" "Cálculo II" "entre 4 y 10"

enviar "Cálculo II" regular 11
check_aviso "nota mayor a 10" "Cálculo II" "entre 4 y 10"

enviar "Cálculo II" regular abc
check_aviso "nota que no es un número" "Cálculo II" "entre 4 y 10"

enviar "Cálculo II" aprobada 8
check_aviso "estado que no existe" "Cálculo II" "no es válido"

request POST /cursadas "materia_id=99999" "estado=cursando" "nota="
check_aviso "materia que no existe" "" "La materia no existe."

request POST /cursadas "materia_id=abc" "estado=cursando" "nota="
check_aviso "id de materia inválido" "" "La materia no es válida."

check_fila "después de los errores, Cálculo II sigue regular con 7.5" "Cálculo II" regular 7.5

section "Correlatividades"
enviar "Redes de Computadoras II" regular 8
check_aviso "regular sin la correlativa" "Redes de Computadoras II" "regulares o finalizadas: Redes de Computadoras I"

enviar "Sistemas Operativos" cursando ""
check_aviso "cursar sin las correlativas" "Sistemas Operativos" "Arquitectura de Computadoras I, Análisis y Diseño de Algoritmos II"

enviar "Análisis y Diseño de Algoritmos I" finalizada 8
check_aviso "finalizar con una correlativa solo regular" "Análisis y Diseño de Algoritmos I" "tener finalizadas: Cálculo I"

enviar "Base de Datos I" cursando ""
check_aviso "cursar con una correlativa cursando" "Base de Datos I" "Análisis y Diseño de Algoritmos I"

check_fila "Redes de Computadoras II no se guardó" "Redes de Computadoras II" sin_cursar
check_fila "Sistemas Operativos no se guardó" "Sistemas Operativos" sin_cursar

enviar "Lenguajes y Paradigmas" sin_cursar ""
check_redirect "sin cursar no exige correlativas: 303 a /" "/"

section "Bajar una materia de la que dependen otras"
enviar "Introducción a la Programación I" sin_cursar ""
check_aviso "Programación I a sin cursar con Programación II finalizada" "Introducción a la Programación I" "dependen de ella: Introducción a la Programación II (finalizada)"

enviar "Introducción a la Programación I" regular 8
check_aviso "Programación I a regular con Programación II finalizada" "Introducción a la Programación I" "Introducción a la Programación II (finalizada)"

enviar "Introducción al Álgebra" cursando ""
check_aviso "Álgebra a cursando con varias dependientes" "Introducción al Álgebra" "Diseño Lógico (finalizada)"

check_fila "Programación I sigue finalizada con 8" "Introducción a la Programación I" finalizada 8

enviar "Programación Orientada a Objetos" sin_cursar ""
check_redirect "bajar una materia sin dependientes cargadas: 303 a /" "/"
check_fila "POO quedó sin cursar" "Programación Orientada a Objetos" sin_cursar

section "Rutas y seguridad"
request GET /cursadas
check "GET /cursadas no está permitido" 405

request GET /no-existe
check "ruta inexistente" 404

request GET "/?materia=Prueba&error=%3Cscript%3Ealert(1)%3C%2Fscript%3E"
check "el aviso escapa el HTML recibido en la URL" 200 "&lt;script&gt;alert(1)&lt;/script&gt;"
check_not "no inyecta la etiqueta <script>" "<script>alert(1)</script>"

# --- Resumen -----------------------------------------------------------------

echo
echo "Resultado: $PASSED verificaciones correctas, $FAILED fallidas."
[ "$FAILED" -eq 0 ]
