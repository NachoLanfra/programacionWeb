# Programación Web - Trabajo Práctico 2

Plan de estudios para la carrera de Ingeniería en Sistemas (UNICEN). El proyecto modela usuarios, materias, 
el progreso de cada usuario en cada materia, y las correlatividades entre materias.

## Requisitos previos

- **Go 1.22**
- **PostgreSQL 15** (vía Docker)
- **sqlc** para generar código Go tipado a partir de SQL plano
- **pgx/v5** como driver de PostgreSQL
- **Docker Compose** para levantar la base de datos

- [Docker](https://docs.docker.com/get-docker/) y Docker Compose
- [Go 1.22+](https://go.dev/dl/)
- [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html) instalado y disponible en el `PATH`

> **Importante:** el daemon de Docker tiene que estar corriendo antes de
> ejecutar `make test`. En Linux con systemd, iniciarlo con:
> ```bash
> sudo systemctl start docker
> ```
> Podés verificar que esté activo con `docker info` (si devuelve
> información del sistema en vez de un error de conexión, está listo).
> En macOS/Windows con Docker Desktop, alcanza con abrir la aplicación y
> esperar a que quede en estado "running".

## Cómo ejecutar los tests

Todo el flujo de testing (levantar la base, generar el código, compilar,
correr los tests y limpiar todo al final) está automatizado. Alcanza con:

```bash
git clone https://github.com/NachoLanfra/programacionWeb.git
cd programacionWeb
git checkout tp2
make test
```

`make test` (que internamente corre `test.sh`) hace lo siguiente:

1. **Tareas previas:**
   - Baja contenedores y volúmenes que hayan quedado de una corrida anterior (`make clean`).
   - Genera el código de acceso a datos con sqlc (`make generate`).
   - Compila el proyecto (`make build`).
   - Levanta la base de datos con Docker Compose (`make docker`).
   - Espera (con timeout) a que Postgres esté listo para aceptar conexiones (`pg_isready`).
2. **Ejecución de tests:** corre `go test ./... -v`, que ejercita todo el paquete `testing` de Go contra la base real.
3. **Tareas posteriores:** al salir del script (haya pasado lo que haya pasado), se bajan de nuevo los contenedores y volúmenes (`make clean`), para no dejar residuos.

## Cómo correr el proyecto localmente (fuera de los tests)

1. Copiar el archivo de variables de entorno de ejemplo:
   ```bash
   cp .env.example .env
   ```
   (Podés editar los valores de usuario/contraseña/nombre de base si querés otros.)

2. Levantar la base de datos:
   ```bash
   make docker
   ```

3. Generar el código de acceso a datos y compilar:
   ```bash
   make generate
   make build
   ```

4. Correr el binario generado (sirve la página estática en `http://localhost:8080`):
   ```bash
   ./tmp/programacion-web
   ```

5. Cuando termines, bajar la base y limpiar:
   ```bash
   make clean
   ```

## Persistencia

### Motor de base de datos

El proyecto usa **PostgreSQL 15**, corriendo en un contenedor Docker.
Al levantar el contenedor por primera vez (o después de
un `make clean`, que borra el volumen), el esquema se inicializa
automáticamente montando `db/schema/schema.sql` en
`/docker-entrypoint-initdb.d/schema.sql`, que Postgres ejecuta en el primer
arranque.

### Modelo de datos

| Tabla                     | Descripción                                                                                                    |
|----------------------------|------------------------------------------------------------------------------------------------------------------|
| `usuario`                  | Alumnos de la carrera: nombre y apellido, email (único) y teléfono.                                             |
| `materia`                  | Materias del plan de estudios: nombre, año y cuatrimestre en que se dicta.                                      |
| `usuario_materia`           | Tabla intermedia (N a N) entre `usuario` y `materia`: guarda el **estado** de cursada de cada alumno en cada materia (por ejemplo `cursando`, `aprobada`) y, opcionalmente, la **nota**. |
| `correlatividad_materia`    | Tabla intermedia (N a N) de `materia` consigo misma: indica qué materias son correlativas (requisito) de cada materia. |

Todas las relaciones tienen `ON DELETE CASCADE`: si se borra un `usuario` o
una `materia`, se borran automáticamente las filas asociadas en las tablas
intermedias, evitando referencias huérfanas.

### Acceso a datos

- **sqlc** genera código Go tipado (`db/sqlc/*.go`) a partir de:
  - el esquema en `db/schema/schema.sql`, y
  - las queries SQL escritas a mano en `db/queries/*.sql` (con anotaciones `-- name: ... :one/:many/:exec`).

  Este código generado **no se versiona** (está en `.gitignore`); se regenera
  con `sqlc generate` (o `make generate`) como parte del flujo de `make test`.

- **pgx/v5** es el driver de PostgreSQL que usa el código generado para
  ejecutar las queries contra la base.

- La conexión se arma con una cadena `postgres://usuario:password@host:puerto/basededatos`,
  usando las variables de entorno descriptas arriba.

### Tests de persistencia

Los tests en `db/tests/` son **tests de integración**: se conectan a una base
PostgreSQL real (levantada por Docker como parte de `make test`) y ejercitan
el ciclo CRUD completo de cada entidad usando el paquete `testing` de Go:

- `usuario_test.go`: alta, lectura, listado, actualización y baja de usuarios.
- `materia_test.go`: alta, lectura, listado, actualización y baja de materias.
- `usuario_materia_test.go`: asociación de un usuario a una materia (estado y nota), lectura del progreso, actualización y baja.
- `correlatividad_materia_test.go`: alta de correlatividades entre materias, listado, actualización y baja, incluyendo la verificación del borrado en cascada.

Cada test limpia los datos que crea (`t.Cleanup`), y además `test.sh` baja el
volumen de Postgres al finalizar, así que cada corrida de `make test` arranca
desde una base limpia.
