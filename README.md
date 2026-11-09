# Programación Web - Trabajo Práctico 5: Renderizado del lado del servidor con templ

Plan de estudios de Ingeniería de Sistemas (UNICEN, plan 2023). La página muestra las 38 materias del plan y deja
que el alumno registre, para cada una, su **estado** (sin cursar, cursando, regular o finalizada) y su **nota**,
respetando las **correlatividades** entre materias.

En el **TP5** se abandona el esquema API JSON + cliente JavaScript del TP4: el servidor Go consulta la base,
arma el HTML completo con componentes [templ](https://templ.guide) y lo envía ya renderizado. El navegador no
ejecuta JavaScript propio; los formularios hacen un `POST` clásico y el servidor responde con una redirección
(patrón **Post/Redirect/Get**).

## Requisitos previos

- [Docker](https://docs.docker.com/get-docker/) y Docker Compose (v2: `docker compose`)
- [Go](https://go.dev/dl/) en la versión que indica `go.mod` (1.26)
- [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html) disponible en el `PATH`
- templ: `go install github.com/a-h/templ/cmd/templ@latest` (el Makefile lo busca en `$(go env GOPATH)/bin`)
- `make`

> El daemon de Docker tiene que estar corriendo antes de ejecutar `make run-app`.
> En Linux con systemd: `sudo systemctl start docker` (verificar con `docker info`).

## Cómo ejecutar

```bash
git clone https://github.com/NachoLanfra/programacionWeb.git
cd programacionWeb
git checkout tp5
make run-app
```

y abrir <http://localhost:8080>. Para detener la aplicación, `Ctrl+C`. Después, para borrar contenedores y volúmenes:

```bash
make clean
```

`make run-app` hace todo el flujo:

1. **Tareas previas**
   - Borra contenedores y volúmenes de una corrida anterior (`make clean`).
   - Genera el código: `templ generate` (vistas) y `sqlc generate` (queries).
   - Compila el proyecto en `tmp/` (`make build`).
   - Levanta PostgreSQL con Docker Compose y espera (con timeout de 60 s) a que acepte conexiones.
     Como el volumen es nuevo, Postgres crea el esquema y carga la seed automáticamente.
2. **Ejecuta la aplicación** en el puerto 8080, hasta que se detiene con `Ctrl+C`.
3. **Tareas posteriores:** `make clean` borra contenedores y volúmenes.

### Targets del Makefile

| Comando         | Qué hace                                                                          |
|-----------------|------------------------------------------------------------------------------------|
| `make run-app`  | **Ciclo completo del TP5:** limpia, genera, compila, levanta la base y ejecuta.    |
| `make run`      | Uso diario: genera, levanta la base **sin borrar datos** y arranca con `go run`.   |
| `make generate` | Ejecuta `templ generate` y `sqlc generate`.                                        |
| `make templ`    | Solo `templ generate`.                                                             |
| `make build`    | Genera el código y compila el binario en `tmp/`.                                   |
| `make db-up`    | Levanta PostgreSQL y espera a que esté listo.                                      |
| `make db-down`  | Baja los contenedores (los datos se conservan en el volumen).                      |
| `make seed`     | Carga el plan de estudios en una base que ya existe (ver [Seed](#seed)).           |
| `make test`     | Tests de persistencia del TP2 (`go test ./...` contra PostgreSQL real).            |
| `make clean`    | Baja contenedores, **borra el volumen de la base** y borra `tmp/`.                 |

## Cómo funciona la página

### Rutas

| Método | Ruta        | Qué hace                                                                                       |
|--------|-------------|-------------------------------------------------------------------------------------------------|
| `GET`  | `/`         | Lista el plan del alumno (materias + su estado y nota) y lo renderiza con templ.                |
| `POST` | `/cursadas` | Recibe el formulario de una fila, valida, guarda y redirige a `/` (303 See Other).              |
| `GET`  | `/health`   | Responde `ok`.                                                                                  |

Si el `POST` no pasa alguna validación, el servidor redirige a `/?materia=...&error=...` y la página muestra el
aviso arriba de la tabla. Sigue siendo PRG: el navegador siempre termina en un `GET`, así que recargar con F5
no reenvía el formulario.

### Componentes templ (`views/`)

| Archivo               | Componente           | Qué renderiza                                                                    |
|-----------------------|----------------------|-----------------------------------------------------------------------------------|
| `layout.templ`        | `Layout(titulo)`     | Estructura base (`html`, `head`, `body`), con [Pico.css](https://picocss.com) por CDN. |
| `cursada_list.templ`  | `CursadaList(plan)`  | Tabla con todas las materias del plan, ordenadas por año y cuatrimestre.          |
| `cursada_form.templ`  | `CursadaForm(fila)`  | Formulario de cada fila: estado, nota y botón Guardar (`POST /cursadas`).         |
| `index.templ`         | `IndexPage(...)`     | Página principal: usa el layout, muestra el aviso de error (si hay) y la tabla.   |

`templ generate` compila cada `.templ` a un `*_templ.go`. Esos archivos se versionan (el Dockerfile no corre
templ) y no se editan a mano.

### Alumno

No hay login: la página muestra siempre el plan de un **alumno de prueba** (`alumno@prueba.com`) que crea la
seed. El servidor lo busca por email al arrancar; si no existe, termina con un mensaje que sugiere correr `make seed`.

### Reglas de validación

**Estado y nota:**

| Estado      | Nota                              |
|-------------|------------------------------------|
| Sin cursar  | No lleva (se rechaza si viene).    |
| Cursando    | No lleva (se rechaza si viene).    |
| Regular     | Obligatoria, entre 4 y 10.         |
| Finalizada  | Obligatoria, entre 4 y 10. Incluye promoción o final rendido. |

**Correlatividades:** cada estado exige algo distinto de las correlativas de la materia:

| Para marcar una materia como... | Sus correlativas tienen que estar...  |
|---------------------------------|----------------------------------------|
| Sin cursar                      | En cualquier estado (no se exige nada). |
| Cursando o Regular              | Regulares o finalizadas.               |
| Finalizada                      | Finalizadas.                           |

Por ejemplo, para cursar Introducción a la Programación II hay que tener Introducción a la Programación I
regular o finalizada; para finalizarla, Introducción a la Programación I tiene que estar finalizada.

La regla también se controla al revés: no se puede **bajar** el estado de una materia si otra que depende de
ella dejaría de cumplir sus correlativas. Por ejemplo, si Introducción a la Programación II está finalizada,
Introducción a la Programación I no puede volver a regular.

Todas las validaciones se hacen en el servidor (`handlers.go`). Los atributos del HTML (`min`, `max`) solo
ayudan a avisar antes de enviar.

## Persistencia

### Motor de base de datos

**PostgreSQL 15** en un contenedor Docker. Al crear la base por primera vez (o después de `make clean`),
Postgres ejecuta en orden los scripts montados en `/docker-entrypoint-initdb.d/`:

1. `01-schema.sql` ← `db/schema/schema.sql` (tablas y restricciones).
2. `02-seed.sql` ← `db/seed/seed.sql` (plan de estudios y alumno de prueba).

### Modelo de datos

| Tabla                    | Descripción                                                                                       |
|--------------------------|----------------------------------------------------------------------------------------------------|
| `usuario`                | Alumnos: nombre y apellido, email (único) y teléfono.                                              |
| `materia`                | Materias del plan: nombre (único), año y cuatrimestre.                                             |
| `usuario_materia`        | Cursada de un alumno en una materia: **estado** y **nota** (opcional).                             |
| `correlatividad_materia` | Qué materias son requisito de cada materia (relación N a N de `materia` consigo misma).            |

Todas las relaciones tienen `ON DELETE CASCADE`: si se borra un usuario o una materia, se borran sus filas
en las tablas intermedias.

### Seed

`db/seed/seed.sql` carga:

- las **38 materias** del plan 2023 y sus **52 correlatividades** (buscadas por nombre, no por id);
- el **alumno de prueba**;
- sus **cursadas iniciales**: un alumno de 2° año, con 1° año finalizado o regular y las materias del
  primer cuatrimestre de 2° cursando. Respetan las correlatividades.

`make seed` se puede correr **varias veces sin duplicar datos**: todos los `INSERT` usan
`ON CONFLICT DO NOTHING`, que saltea las filas que ya existen. Por lo mismo, **no pisa** lo que ya está cargado: si se cambiaron cursadas desde la
página, `make seed` no las vuelve a su valor original. Para volver al estado inicial:

```bash
make clean   # borra el volumen
make run     # recrea la base: esquema + seed
```

### Acceso a datos

- **sqlc** genera código Go tipado (`db/sqlc/`) a partir del esquema y de las queries de `db/queries/*.sql`.
  Ese código **no se versiona** (está en `.gitignore`); se regenera con `make generate`, que corre como parte
  de `make run`, `make build` y `make test`.
- **pgx/v5** es el driver de PostgreSQL; el servidor usa un pool de conexiones (`pgxpool`).

### Credenciales de la base

No hace falta crear un `.env` para correr el proyecto. Si no existe (o las variables quedan vacías),
`docker-compose.yml`, el Makefile y la aplicación usan los mismos valores por defecto
(`postgres` / `postgres` / `programacion_web`). `.env.example` lista las variables que se pueden definir
(`POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`).

Esto es solo para el entorno del TP. En un proyecto real no dejaríamos credenciales por defecto y la
aplicación debería fallar si falta alguna.

### Tests de persistencia (TP2)

Los tests de `db/tests/` son **tests de integración**: se conectan a una base PostgreSQL real y ejercitan el
CRUD de cada tabla. Se corren con `make test`.

## Trabajos anteriores

Como pide la consigna, en el TP5 se eliminaron los handlers de la API JSON. La API REST está en la rama `tp3`
y el frontend con HTML/JS/CSS en la rama `tp4`.
