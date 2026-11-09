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
	queries  *sqlc.Queries
	alumnoID int32 // alumno fijo cuya planilla muestra la página
}

// emailAlumno identifica al alumno de prueba que crea la seed.
const emailAlumno = "alumno@prueba.com"

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
	queries := sqlc.New(pool)
	alumno, err := queries.GetUsuarioByEmail(ctx, emailAlumno)
	if err != nil {
		log.Fatalf("no se encontró el alumno %s (¿corriste make seed?): %v", emailAlumno, err)
	}
	srv := &Server{queries: queries, alumnoID: alumno.IDUsuario}

	// Rutas
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// Página principal renderizada en el servidor
	mux.HandleFunc("GET /{$}", srv.index)
	mux.HandleFunc("POST /cursadas", srv.guardarCursada)

	// Arrancar
	serverPort := ":8080"
	log.Printf("Servidor escuchando en %s", serverPort)
	if err := http.ListenAndServe(serverPort, mux); err != nil {
		log.Fatalf("Error al iniciar el servidor: %s", err)
	}
}
