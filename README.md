# Programación Web - Trabajo Práctico 2

## Requisitos previos
* [Docker](https://docs.docker.com/get-docker/) y Docker Compose instalados.

## Pasos para ejecutar localmente

1. **Clonar el repositorio:**
   ```bash
   git clone https://github.com/NachoLanfra/programacionWeb.git
   ```

2. **Ingresar al directorio:**
   ```bash
   cd programacionWeb
   ```

3. **Posicionarse en la rama del TP2:**
   ```bash
   git checkout tp2
   ```

4. **Configurar las variables de entorno:**
   Crea y abre el archivo `.env` en la raíz del proyecto:
   ```bash
   micro .env
   ```
   Pega el siguiente contenido dentro del archivo y guárdalo para que la base de datos PostgreSQL pueda inicializarse:
   ```env
   POSTGRES_USER=postgres
   POSTGRES_PASSWORD=contraseña_de_la_bd
   POSTGRES_DB=nombre_de_la_bd
   ```

5. **Levantar el proyecto:**
   ```bash
   docker compose up -d
   ```
