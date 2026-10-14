package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"

	"ejerEsp.com/ejerEsp/db/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	queries *sqlc.Queries
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {

	// 1. Config
	user := getEnvOrDefault("POSTGRES_USER", "postgres")
	pass := getEnvOrDefault("POSTGRES_PASSWORD", "")
	name := getEnvOrDefault("POSTGRES_DB", "plan_estudios")
	host := getEnvOrDefault("DB_HOST", "localhost")
	dbPort := getEnvOrDefault("DB_PORT", "5432")

	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, pass),
		Host:   host + ":" + dbPort,
		Path:   name,
	}
	dsn := u.String()

	// 2. Pool
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("no se pudo crear el pool: %v", err)
	}
	defer pool.Close()

	// 3. Verificar conexión
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("no se pudo conectar a la base: %v", err)
	}
	log.Println("Conectado a la base de datos")

	// 4. Server con las queries de sqlc
	srv := &Server{queries: sqlc.New(pool)}

	// 5. Rutas
	mux := http.NewServeMux()

	// Servidor de estáticos
	staticDir := "./static"
	fileServer := http.FileServer(http.Dir(staticDir))
	mux.Handle("/", fileServer)

	// API de Materia
	mux.HandleFunc("POST /materias", srv.crearMateria)
	mux.HandleFunc("GET /materias", srv.listarMaterias)
	mux.HandleFunc("GET /materias/{id}", srv.obtenerMateria)
	mux.HandleFunc("PUT /materias/{id}", srv.actualizarMateria)
	mux.HandleFunc("DELETE /materias/{id}", srv.eliminarMateria)

	// API de Correlatividades de Materia
	mux.HandleFunc("POST /materias/{id}/correlativas", srv.createCorrelativa)
	mux.HandleFunc("GET /materias/{id}/correlativas", srv.listCorrelativas)
	mux.HandleFunc("DELETE /materias/{id}/correlativas/{requeridaId}", srv.deleteCorrelativa)

	// API de Usuario
	mux.HandleFunc("POST /usuarios", srv.createUsuario)
	mux.HandleFunc("GET /usuarios", srv.listUsuarios)
	mux.HandleFunc("GET /usuarios/{id}", srv.getUsuario)
	mux.HandleFunc("PUT /usuarios/{id}", srv.updateUsuario)
	mux.HandleFunc("DELETE /usuarios/{id}", srv.deleteUsuario)

	// API de Usuario-Materia
	mux.HandleFunc("POST /usuarios/{id}/materias", srv.createUsuarioMateria)
	mux.HandleFunc("GET /usuarios/{id}/materias", srv.listUsuarioMateria)
	mux.HandleFunc("GET /usuarios/{id}/materias/{materiaId}", srv.getUsuarioMateria)
	mux.HandleFunc("PUT /usuarios/{id}/materias/{materiaId}", srv.updateUsuarioMateria)
	mux.HandleFunc("DELETE /usuarios/{id}/materias/{materiaId}", srv.deleteUsuarioMateria)

	// 6. Arrancar
	serverPort := ":8080"
	log.Printf("Servidor escuchando en %s", serverPort)
	if err := http.ListenAndServe(serverPort, mux); err != nil {
		log.Fatalf("Error al iniciar el servidor: %s", err)
	}
}
