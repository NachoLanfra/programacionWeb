# Programación Web - Trabajo Práctico 6: Interfaces reactivas con HTMX

Plan de estudios de Ingeniería de Sistemas (UNICEN, plan 2023). La página muestra las 38 materias del plan y deja
que el alumno registre, para cada una, su **estado** (sin cursar, cursando, regular o finalizada) y su **nota**,
respetando las **correlatividades** entre materias.

El servidor Go consulta la base y arma el HTML con componentes [templ](https://templ.guide) (TP5). En el **TP6**
se agrega [HTMX](https://htmx.org) para que la página se comporte como una aplicación de página única: guardar y
borrar una cursada actualiza solo la parte afectada de la página, **sin recargarla y sin escribir JavaScript**.
El servidor responde fragmentos de HTML en lugar de la página completa.

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
git checkout tp6
make run-app
```

y abrir <http://localhost:8080>. Para detener la aplicación, `Ctrl+C`: la limpieza se hace sola.

`make run-app` ejecuta `run-app.sh`, que hace todo el flujo:

1. **Tareas previas**
   - Borra contenedores y volúmenes de una corrida anterior (`make clean`).
   - Genera el código: `templ generate` (vistas) y `sqlc generate` (queries).
   - Compila el proyecto en `tmp/` (`make build`).
   - Levanta PostgreSQL con Docker Compose y espera (con timeout de 60 s) a que acepte conexiones.
     Como el volumen es nuevo, Postgres crea el esquema y carga la seed automáticamente.
2. **Ejecuta la aplicación** en el puerto 8080, hasta que se detiene con `Ctrl+C`.
3. **Tareas posteriores:** detiene el servidor y borra contenedores y volúmenes (`make clean`).
   Corren **siempre**, tanto al terminar con `Ctrl+C` como si falla algún paso anterior (por ejemplo, la
   compilación), gracias a un `trap` de bash. Si algo falló, el script termina con código de error.

### Targets del Makefile

| Comando         | Qué hace                                                                          |
|-----------------|------------------------------------------------------------------------------------|
| `make run-app`  | **Ciclo completo:** limpia, genera, compila, levanta la base, ejecuta y limpia al salir. |
| `make run`      | Uso diario: genera, levanta la base **sin borrar datos** y arranca con `go run`.   |
| `make generate` | Ejecuta `templ generate` y `sqlc generate`.                                        |
| `make templ`    | Solo `templ generate`.                                                             |
| `make build`    | Genera el código y compila el binario en `tmp/`.                                   |
| `make db-up`    | Levanta PostgreSQL y espera a que esté listo.                                      |
| `make db-down`  | Baja los contenedores (los datos se conservan en el volumen).                      |
| `make seed`     | Carga el plan de estudios en una base que ya existe (ver [Seed](#seed)).           |
| `make test-html`| Pruebas automáticas del recorrido HTML (ver [Pruebas](#pruebas)).                  |
| `make test`     | Tests de persistencia del TP2 (`go test ./...` contra PostgreSQL real).            |
| `make clean`    | Baja contenedores, **borra el volumen de la base** y borra `tmp/`.                 |

## Estructura del proyecto

| Archivo / carpeta      | Contenido                                                                          |
|------------------------|-------------------------------------------------------------------------------------|
| `main.go`              | Configuración, conexión a la base (pool de pgx), rutas y arranque del servidor.     |
| `handlers.go`          | Handlers HTTP: página principal, guardar y borrar cursadas, validaciones.           |
| `views/`               | Componentes templ (`.templ`) y el código Go que genera templ (`*_templ.go`).        |
| `db/schema/schema.sql` | Tablas y restricciones de la base.                                                  |
| `db/queries/`          | Queries SQL a partir de las que sqlc genera el código de acceso a datos.            |
| `db/seed/seed.sql`     | Plan de estudios, alumno de prueba y cursadas iniciales.                            |
| `db/tests/`            | Tests de persistencia (TP2).                                                        |
| `run-app.sh`           | Ciclo completo para ejecutar la aplicación (`make run-app`).                        |
| `test-html.sh`         | Ciclo completo de las pruebas del recorrido HTML (`make test-html`).                |
| `requests-html.sh`     | Las pruebas del recorrido HTML, con `curl`.                                         |
| `Makefile`             | Tareas: generar código, compilar, levantar la base, ejecutar, probar y limpiar.     |
| `docker-compose.yml`   | PostgreSQL (y la aplicación, si se quiere correr en Docker).                        |

## Cómo funciona la página

### Rutas

| Método   | Ruta                     | Qué hace                                                                                  |
|----------|--------------------------|--------------------------------------------------------------------------------------------|
| `GET`    | `/`                      | Página completa: el plan del alumno (materias con su estado y nota).                      |
| `POST`   | `/cursadas`              | Guarda la cursada de una fila. Con HTMX devuelve la tabla; sin HTMX redirige a `/`.        |
| `DELETE` | `/cursadas/{materiaId}`  | Borra la cursada de una materia y devuelve su fila como "sin cursar". Solo con HTMX.       |
| `GET`    | `/health`                | Responde `ok`.                                                                            |

### Interactividad con HTMX

HTMX se carga desde un CDN en el `<head>` del layout (versión 4.0.0, con `integrity`). Toda la interactividad se declara con atributos `hx-*` en el HTML.

El servidor distingue los pedidos de HTMX por el header `HX-Request: true`.

**Guardar (`POST /cursadas`)**

Cada fila tiene un formulario con `hx-post="/cursadas"`, `hx-target="#plan"` y `hx-swap="outerHTML"`:

1. HTMX envía el formulario en segundo plano.
2. El servidor valida y guarda. En lugar de redirigir, vuelve a consultar el plan y devuelve **dos fragmentos**:
   - `<section id="plan">`: la tabla actualizada, que reemplaza a la anterior (`hx-target`).
   - `<div id="aviso" hx-swap-oob="true">`: el aviso de error. Con `hx-swap-oob` ("out of band") HTMX lo
     reemplaza aparte de la tabla. Si el guardado salió bien llega vacío y borra un error anterior.
3. Si hay un error, la respuesta es la misma pero con el aviso lleno y el código HTTP del error (422, 400, 404 o
   500). HTMX 4 también reemplaza el contenido de las respuestas 4xx/5xx.

El formulario conserva `method="POST"` y `action="/cursadas"`: si HTMX no carga, se envía de forma tradicional y el
servidor responde con **Post/Redirect/Get** (el error viaja en la URL: `/?materia=...&error=...`).

**Borrar (`DELETE /cursadas/{materiaId}`)**

Las filas con algo cargado tienen un botón **Borrar** con `hx-delete="/cursadas/{id}"`, `hx-target="closest tr"`,
`hx-swap="outerHTML"` y `hx-confirm` (pide confirmación antes de enviar).

- **Qué se borra:** la **cursada** del alumno (la fila de `usuario_materia`), no la materia. Borrar una materia
  no tendría sentido en esta aplicación: las materias son el plan de estudios, vienen cargadas y el alumno no las
  crea ni las elimina. Lo que el alumno sí carga, y por lo tanto puede borrar, es su estado y su nota en cada una.
- **Diferencia con la consigna:** la consigna pide que el servidor devuelva una **respuesta vacía** con 200, para
  que la fila desaparezca de la tabla. En nuestro caso eso dejaría la página inconsistente con el plan de estudios: después de borrar la cursada la
  materia sigue en el plan, así que la fila desaparecería y volvería a aparecer, como "sin cursar", al recargar.
  Por eso nos apartamos de ese punto a propósito: el servidor devuelve la **fila actualizada**, como "sin cursar" y
  sin botón Borrar. Se mantiene todo lo demás que pide la consigna (`hx-delete`, `closest tr`, `outerHTML`,
  `hx-confirm` y el método `Delete...` de sqlc); solo cambia el contenido de la respuesta, para que la página
  muestre siempre lo que hay en la base.
- **Validación:** borrar equivale a pasar la materia a "sin cursar", así que se rechaza (422) si otra materia
  cargada depende de ella. La fila queda como estaba y se muestra el aviso.
- **El aviso:** viaja dentro de un `<hx-partial hx-target="#aviso">`, la alternativa de HTMX 4 a `hx-swap-oob`. Se
  usa porque la respuesta empieza con un `<tr>`, y el navegador descarta un `<div>` suelto al lado de una fila de
  tabla; `<hx-partial>` se convierte en un `<template>`, que sí puede ir ahí.
- **Materia inexistente o id inválido:** el servidor responde con los headers `HX-Retarget: #aviso` y
  `HX-Reswap: innerHTML`, para que el error se muestre en el aviso y no reemplace la fila.

El borrado solo funciona con HTMX: un formulario HTML solo puede enviar `GET` o `POST`, no `DELETE`.

### Componentes templ (`views/`)

| Archivo               | Componente           | Qué renderiza                                                                    |
|-----------------------|----------------------|-----------------------------------------------------------------------------------|
| `layout.templ`        | `Layout(titulo)`           | Estructura base, con [Pico.css](https://picocss.com) y HTMX por CDN.               |
| `index.templ`         | `IndexPage(...)`           | Página principal: el alumno, el aviso y la tabla.                                  |
| `cursada_list.templ`  | `CursadaList(plan)`        | Tabla `#plan` con las 38 materias. Es el fragmento que devuelve el guardado.       |
| `cursada_list.templ`  | `CursadaFila(fila)`        | Una fila: datos de la materia, formulario y botón Borrar.                          |
| `cursada_list.templ`  | `FilaConAviso(...)`        | Respuesta del borrado: la fila y el aviso dentro de un `<hx-partial>`.             |
| `cursada_form.templ`  | `CursadaForm(fila)`        | Formulario de cada fila: estado, nota y botón Guardar (`hx-post`).                 |
| `aviso.templ`         | `AvisoBox(aviso, oob)`     | Contenedor `#aviso`; con `oob` se marca con `hx-swap-oob`.                          |

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

## Pruebas

Hay dos niveles de pruebas automáticas. Ambos levantan una base nueva y limpian contenedores y volúmenes al terminar.

### Pruebas del recorrido HTML

```bash
make test-html
```

Prueban la aplicación de punta a punta, como la usaría el navegador. `test-html.sh` hace el ciclo completo:

1. **Tareas previas:** `make clean`, genera el código, compila y levanta PostgreSQL (con el esquema y la seed).
2. **Prueba:** levanta el servidor, espera a que `GET /health` responda y ejecuta `requests-html.sh`.
3. **Tareas posteriores:** detiene el servidor y corre `make clean`, siempre (gracias a un `trap`). Si algo falló,
   antes muestra los últimos logs del servidor.

`requests-html.sh` hace las peticiones con `curl`: pide la página con `GET` y manda los formularios con `POST`
(como `application/x-www-form-urlencoded`), **sin seguir las redirecciones**. Prueba los dos caminos: sin HTMX
(Post/Redirect/Get) y con HTMX (agregando el header `HX-Request: true`, como lo hace HTMX). Verifica:

- **Página principal:** responde 200, usa el layout con Pico.css, muestra el alumno y las 38 materias.
- **Datos de la seed:** el estado seleccionado y la nota de varias filas.
- **Guardado válido:** el `POST /cursadas` responde 303 a `/` y, al volver a pedir la página, la fila muestra el
  estado y la nota nuevos (incluida una nota decimal).
- **Validaciones de estado y nota:** cada caso inválido (nota en sin cursar o cursando, regular o finalizada sin
  nota, nota fuera de 4 a 10 o no numérica, estado o materia inexistentes) redirige con el error y la página
  muestra el aviso con la materia y el mensaje correctos.
- **Correlatividades:** se rechaza cursar, regularizar o finalizar sin las correlativas necesarias, y bajar una
  materia de la que dependen otras ya cargadas. En todos los casos se verifica que **no se guardó nada**.
- **Guardar con HTMX:** la página carga HTMX y el formulario tiene `hx-post`, `hx-target` y `hx-swap`. Un guardado
  válido responde 200 con **solo** la tabla (sin `<html>`) y el aviso vacío por `hx-swap-oob`; uno inválido
  responde 422 con la tabla y el aviso. Se verifica en la base qué quedó guardado.
- **Borrar con HTMX:** solo las filas cargadas tienen el botón (con `hx-delete`, `closest tr`, `outerHTML` y
  `hx-confirm`). Un borrado válido responde 200 con solo la fila, ahora "sin cursar", y la cursada desaparece de la
  base; con dependientes responde 422, la fila queda igual y la base no cambia; una materia inexistente da 404 con
  `HX-Retarget: #aviso` y un id inválido, 400.
- **Rutas y seguridad:** `GET /cursadas` da 405, una ruta inexistente da 404, y el aviso escapa el HTML recibido en
  la URL (no se puede inyectar un `<script>`).

La salida termina con una línea como `Resultado: 127 verificaciones correctas, 0 fallidas.` Si alguna falla, el
script termina con código distinto de 0.

### Tests de persistencia (TP2)

```bash
make test
```

Los tests de `db/tests/` son **tests de integración**: se conectan a una base PostgreSQL real y ejercitan el
CRUD de cada tabla con el paquete `testing` de Go.

## Trabajos anteriores

- **TP3:** API REST en JSON, en la rama `tp3`. Se eliminó en el TP5, como pedía la consigna.
- **TP4:** frontend con HTML, JavaScript y CSS que consumía la API, en la rama `tp4`.
- **TP5:** renderizado en el servidor con templ y formularios con Post/Redirect/Get, en la rama `tp5`.
