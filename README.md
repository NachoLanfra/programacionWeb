# Programación Web - Trabajo Práctico 3: API REST

Plan de estudios para la carrera de Ingeniería en Sistemas (UNICEN). El proyecto modela usuarios, materias,
el progreso de cada usuario en cada materia, y las correlatividades entre materias.

En el **TP3** la capa de datos del TP2 (PostgreSQL + sqlc) se expone como una **API REST** en JSON, implementada
con la biblioteca estándar de Go (`net/http`, con las rutas por método y variables de Go 1.22+) y empaquetada con Docker.

## Requisitos previos

- [Docker](https://docs.docker.com/get-docker/) y Docker Compose (v2: `docker compose`)
- [Go](https://go.dev/dl/) en la versión que indica `go.mod` (1.25)
- [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html) instalado y disponible en el `PATH`
- `make` y `curl`

> **Importante:** el daemon de Docker tiene que estar corriendo antes de ejecutar `make test-api`.
> En Linux con systemd: `sudo systemctl start docker` (verificar con `docker info`).
> En macOS/Windows con Docker Desktop alcanza con abrir la aplicación y esperar a que quede en "running".

## Cómo ejecutar y probar la API

Todo el flujo (generar código, compilar, levantar la base y la app, probar cada endpoint y limpiar) está automatizado:

```bash
git clone https://github.com/NachoLanfra/programacionWeb.git
cd programacionWeb
git checkout tp3
make test-api
```

`make test-api` (que corre `test-api.sh`) hace lo siguiente:

1. **Tareas previas**
   - Baja contenedores y volúmenes que hayan quedado de una corrida anterior (`make clean`).
   - Genera el código de acceso a datos con sqlc (`make generate`).
   - Compila el proyecto (`make build`).
   - Levanta PostgreSQL con Docker Compose y espera (con timeout) a que acepte conexiones (`pg_isready`).
   - Construye la imagen de la aplicación, la levanta y espera (con timeout) a que `GET /health` responda.
2. **Prueba de cada endpoint:** ejecuta `requests.sh`, que hace las peticiones con `curl` y **verifica el código HTTP
   y el contenido** de cada respuesta (casos válidos y de error). Si alguna verificación falla, termina con código distinto de 0.
3. **Tareas posteriores:** al salir (haya pasado lo que haya pasado) se bajan contenedores y volúmenes (`make clean`).
   Si algo falló, antes se imprimen los últimos logs de la aplicación.

La salida termina con una línea como `Resultado: 75 verificaciones correctas, 0 fallidas.`

### Levantar la API para usarla a mano

```bash
make generate                 # genera el código de sqlc (db/sqlc no se versiona)
docker compose up -d --build  # levanta PostgreSQL + la API
./requests.sh                 # (opcional) corre las mismas pruebas contra la API levantada
make clean                    # baja todo y borra el volumen de la base
```

La API queda en `http://localhost:8080` (el puerto del host se cambia con `APP_PORT`) y la página estática del TP1 en `/`.
`requests.sh` acepta otra dirección con `BASE_URL=http://host:puerto ./requests.sh`.

### Targets del Makefile

| Comando         | Qué hace                                                                       |
|-----------------|---------------------------------------------------------------------------------|
| `make test-api` | **Ciclo completo del TP3:** preparar, levantar, probar la API y limpiar.         |
| `make generate` | Ejecuta sqlc.                                                                    |
| `make build`    | Compila el binario en `tmp/`.                                                    |
| `make test`     | Tests de persistencia del TP2 (`go test ./...` contra PostgreSQL real).          |
| `make clean`    | Baja contenedores y volúmenes y borra `tmp/`.                                    |

## API REST

Todas las respuestas son JSON (`Content-Type: application/json`), salvo los `204 No Content`.
Los campos usan los mismos nombres que las columnas de la base (`id_usuario`, `nombre_apellido`, ...).

### Endpoints

| Método   | Ruta                                         | Descripción                                             | Respuestas              |
|----------|----------------------------------------------|---------------------------------------------------------|-------------------------|
| `GET`    | `/health`                                    | Estado de la API                                        | 200                     |
| `POST`   | `/usuarios`                                  | Crea un usuario                                         | 201, 400, 409           |
| `GET`    | `/usuarios`                                  | Lista todos los usuarios                                | 200                     |
| `GET`    | `/usuarios/{id}`                             | Obtiene un usuario                                      | 200, 400, 404           |
| `PUT`    | `/usuarios/{id}`                             | Actualiza un usuario                                    | 200, 400, 404, 409      |
| `DELETE` | `/usuarios/{id}`                             | Elimina un usuario (y su progreso, en cascada)          | 204, 400, 404           |
| `POST`   | `/materias`                                  | Crea una materia                                        | 201, 400                |
| `GET`    | `/materias`                                  | Lista las materias                                      | 200                     |
| `GET`    | `/materias/{id}`                             | Obtiene una materia                                     | 200, 400, 404           |
| `PUT`    | `/materias/{id}`                             | Actualiza una materia                                   | 200, 400, 404           |
| `DELETE` | `/materias/{id}`                             | Elimina una materia (en cascada)                        | 204, 400, 404           |
| `POST`   | `/usuarios/{id}/materias`                    | Registra el estado de un usuario en una materia         | 201, 400, 404, 409, 422 |
| `GET`    | `/usuarios/{id}/materias`                    | Lista el progreso del usuario                           | 200, 400, 404           |
| `GET`    | `/usuarios/{id}/materias/{materiaId}`        | Progreso del usuario en una materia                     | 200, 400, 404           |
| `PUT`    | `/usuarios/{id}/materias/{materiaId}`        | Actualiza estado y nota                                 | 200, 400, 404           |
| `DELETE` | `/usuarios/{id}/materias/{materiaId}`        | Elimina el registro de progreso                         | 204, 400, 404           |
| `GET`    | `/correlatividades`                          | Lista todas las correlatividades                        | 200                     |
| `POST`   | `/materias/{id}/correlativas`                | Agrega una materia requerida a la materia `{id}`        | 201, 400, 404, 409, 422 |
| `GET`    | `/materias/{id}/correlativas`                | Lista las materias requeridas por `{id}`                | 200, 400, 404           |
| `GET`    | `/materias/{id}/correlativas/{requeridaId}`  | Devuelve la materia `{requeridaId}` si `{id}` la requiere | 200, 400, 404         |
| `PUT`    | `/materias/{id}/correlativas/{requeridaId}`  | Pasa la correlatividad a otra materia (ver abajo)       | 200, 400, 404, 409      |
| `DELETE` | `/materias/{id}/correlativas/{requeridaId}`  | Quita una correlatividad                                | 204, 400, 404           |

**`PUT` de una correlatividad:** recibe `{"nueva_materia_id": N}` y cambia la relación "`{id}` requiere `{requeridaId}`"
por "`N` requiere `{requeridaId}`". Verifica que `N` exista, que la correlatividad original exista y que el cambio no forme un ciclo directo.

### Cuerpos de las peticiones y validaciones

| Recurso         | Campos                                    | Validaciones                                                                                    |
|-----------------|-------------------------------------------|--------------------------------------------------------------------------------------------------|
| Usuario         | `nombre_apellido`, `email`, `telefono`    | Obligatorios (se recortan espacios); email con formato válido; máx. 100 / 100 / 20 caracteres.   |
| Materia         | `nombre`, `anio`, `cuatrimestre`          | `nombre` obligatorio; `anio` entre 1 y 5; `cuatrimestre` 1 o 2.                                  |
| Progreso (POST) | `materia_id`, `estado`, `nota` (opcional) | `estado` ∈ `sin_cursar`, `cursando`, `aprobada`, `desaprobada`; `nota` entre 0 y 10 (o ausente). |
| Progreso (PUT)  | `estado`, `nota` (opcional)               | Igual que arriba.                                                                                |
| Correlatividad  | `materia_requerida_id`                    | Entero positivo, distinto de la propia materia; no puede existir la correlatividad inversa.      |

El JSON se lee de forma estricta: se rechazan cuerpos mal formados o con campos desconocidos. Los ids de la URL deben ser enteros positivos.

### Códigos de error

Los errores devuelven `{"error": "mensaje"}`; las validaciones agregan el detalle por campo:

```json
{"error":"error de validación","campos":{"email":"no tiene un formato válido","nombre_apellido":"es obligatorio"}}
```

| Código | Cuándo                                                                                                  |
|--------|----------------------------------------------------------------------------------------------------------|
| 400    | JSON inválido, campos que no pasan la validación, o id de la URL inválido.                               |
| 404    | El recurso no existe (incluye `DELETE` de un id inexistente).                                            |
| 405    | Método no soportado para esa ruta.                                                                       |
| 409    | Conflicto: email repetido, materia ya registrada para el usuario, correlatividad repetida o en ciclo.    |
| 422    | El cuerpo referencia algo que no existe (por ejemplo, una `materia_id` inexistente).                     |
| 500    | Error interno (se registra en el log del servidor sin exponer detalles al cliente).                      |

### Ejemplos con curl

```bash
# Crear (201) -> devuelve el objeto creado y el header Location
curl -i -X POST localhost:8080/usuarios -H 'Content-Type: application/json' \
  -d '{"nombre_apellido":"Ana Pérez","email":"ana@example.com","telefono":"2494111111"}'

curl localhost:8080/usuarios                       # listar
curl localhost:8080/usuarios/1                     # obtener (404 si no existe)
curl -X PUT localhost:8080/usuarios/1 -H 'Content-Type: application/json' \
  -d '{"nombre_apellido":"Ana M. Pérez","email":"ana@example.com","telefono":"2494222222"}'
curl -i -X DELETE localhost:8080/usuarios/1        # 204 No Content

curl -X POST localhost:8080/materias -H 'Content-Type: application/json' \
  -d '{"nombre":"Programación Web","anio":4,"cuatrimestre":1}'

# Progreso del usuario 1 en la materia 1, y correlatividad (la materia 2 requiere la 1)
curl -X POST localhost:8080/usuarios/1/materias -H 'Content-Type: application/json' \
  -d '{"materia_id":1,"estado":"aprobada","nota":8.5}'
curl -X POST localhost:8080/materias/2/correlativas -H 'Content-Type: application/json' \
  -d '{"materia_requerida_id":1}'
```

`requests.sh` contiene las peticiones probadas de cada endpoint principal, con casos válidos e inválidos.


Los handlers reciben el repositorio generado por sqlc (`*sqlc.Queries`) dentro de `Server`. Los errores de PostgreSQL se
traducen a HTTP en un único lugar (`dbError`): sin filas → 404, violación de `UNIQUE` → 409, violación de clave foránea → 422.

## Variables de entorno

| Variable                                            | Uso                                                   | Default                                      |
|-----------------------------------------------------|-------------------------------------------------------|----------------------------------------------|
| `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` | Credenciales de la base (compose, app y tests).       | `postgres` / `postgres` / `programacion_web` |
| `DB_HOST`, `DB_PORT`                                | Dónde está la base (compose fija `DB_HOST=postgres`). | `localhost`, `5432`                          |
| `APP_PORT`                                          | Puerto del host al que se publica la API (compose).   | `8080`                                       |

### Nota sobre las credenciales de la base

`.env.example` solo lista las variables (`POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`) con valor vacío: **no hace
falta completarlo ni crear un `.env`** para correr el proyecto. Alcanza con `make test-api`. Si no existe `.env` (o las
variables quedan vacías), `docker-compose.yml`, la aplicación y los tests en Go caen a los mismos valores por default
(`postgres` / `postgres` / `programacion_web`), pensados exclusivamente para que el flujo de test funcione sin pasos manuales.

Esto es válido solo para este entorno de test, que se crea y se borra en cada corrida y no tiene datos reales. En un
proyecto real no se harían así las cosas: no dejaríamos credenciales por default ni en el código ni en el
`docker-compose.yml`, y la aplicación debería fallar si falta alguna, en vez de arrancar igual con un valor por defecto.

## Persistencia

### Motor de base de datos

El proyecto usa **PostgreSQL 15**, corriendo en un contenedor Docker. Al levantar el contenedor por primera vez
(o después de un `make clean`, que borra el volumen), el esquema se inicializa automáticamente montando
`db/schema/schema.sql` en `/docker-entrypoint-initdb.d/schema.sql`, que Postgres ejecuta en el primer arranque.

### Modelo de datos

| Tabla                    | Descripción                                                                                                   |
|--------------------------|----------------------------------------------------------------------------------------------------------------|
| `usuario`                | Alumnos de la carrera: nombre y apellido, email (único) y teléfono.                                           |
| `materia`                | Materias del plan de estudios: nombre, año y cuatrimestre en que se dicta.                                    |
| `usuario_materia`        | Tabla intermedia (N a N) entre `usuario` y `materia`: guarda el **estado** de cursada y, opcionalmente, la **nota**. |
| `correlatividad_materia` | Tabla intermedia (N a N) de `materia` consigo misma: indica qué materias son requisito de cada materia.       |

Todas las relaciones tienen `ON DELETE CASCADE`: si se borra un `usuario` o una `materia`, se borran
automáticamente las filas asociadas en las tablas intermedias, evitando referencias huérfanas.

### Acceso a datos

- **sqlc** genera código Go tipado (`db/sqlc/*.go`) a partir del esquema (`db/schema/schema.sql`) y de las queries
  escritas a mano en `db/queries/*.sql`. Este código generado **no se versiona** (está en `.gitignore`); se regenera
  con `make generate` como parte de `make test-api` y `make test`.
- **pgx/v5** es el driver de PostgreSQL. La API usa un pool de conexiones (`pgxpool`), porque atiende peticiones concurrentes.

### Tests de persistencia (TP2)

Los tests en `db/tests/` son **tests de integración**: se conectan a una base PostgreSQL real y ejercitan el ciclo
CRUD de cada tabla con el paquete `testing` de Go. Se corren con `make test`. Cada test limpia los datos que crea
(`t.Cleanup`), y `test.sh` baja el volumen de Postgres al finalizar.
