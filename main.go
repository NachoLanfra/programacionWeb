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

	// Config
	user := getEnvOrDefault("POSTGRES_USER", "postgres")
	pass := getEnvOrDefault("POSTGRES_PASSWORD", "postgres")
	name := getEnvOrDefault("POSTGRES_DB", "programacion_web")
	host := getEnvOrDefault("DB_HOST", "localhost")
	dbPort := getEnvOrDefault("DB_PORT", "5432")

	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, pass),
		Host:   host + ":" + dbPort,
		Path:   name,
	}
	dsn := u.String()

	// Pool
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("no se pudo crear el pool: %v", err)
	}
	defer pool.Close()

	// Verificar conexión
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("no se pudo conectar a la base: %v", err)
	}
	log.Println("Conectado a la base de datos")

	// Server con las queries de sqlc
	srv := &Server{queries: sqlc.New(pool)}

	// Rutas
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Servidor de estáticos
	staticDir := "./static"
	fileServer := http.FileServer(http.Dir(staticDir))
	mux.Handle("GET /", fileServer)

	// API de Materia
	mux.HandleFunc("POST /materias", srv.crearMateria)
	mux.HandleFunc("GET /materias", srv.listarMaterias)
	mux.HandleFunc("GET /materias/{id}", srv.obtenerMateria)
	mux.HandleFunc("PUT /materias/{id}", srv.actualizarMateria)
	mux.HandleFunc("DELETE /materias/{id}", srv.eliminarMateria)

	// API de Correlatividades de Materia
	mux.HandleFunc("GET /correlatividades", srv.listTodasCorrelatividades)
	mux.HandleFunc("POST /materias/{id}/correlativas", srv.createCorrelativa)
	mux.HandleFunc("GET /materias/{id}/correlativas", srv.listCorrelativas)
	mux.HandleFunc("GET /materias/{id}/correlativas/{requeridaId}", srv.getCorrelativa)
	mux.HandleFunc("PUT /materias/{id}/correlativas/{requeridaId}", srv.updateCorrelativa)
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

	// Arrancar
	serverPort := ":8080"
	log.Printf("Servidor escuchando en %s", serverPort)
	if err := http.ListenAndServe(serverPort, mux); err != nil {
		log.Fatalf("Error al iniciar el servidor: %s", err)
	}
}
