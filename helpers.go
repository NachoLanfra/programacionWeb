package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// writeJSON envía una respuesta HTTP con código de estado y cuerpo en formato JSON.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("error escribiendo respuesta JSON: %v", err)
	}
}

// writeError envía un error genérico formateado como {"error": msg}.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// writeValidation responde con 400 Bad Request y el mapa de errores por campo.
func writeValidation(w http.ResponseWriter, errs map[string]string) {
	writeJSON(w, http.StatusBadRequest, map[string]any{
		"error":  "error de validación",
		"campos": errs,
	})
}

// decodeJSON decodifica el cuerpo de la petición en la estructura destino.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return false
	}
	return true
}

// pathID extrae y valida un parámetro numérico positivo de la ruta (Go 1.22+).
func pathID(w http.ResponseWriter, r *http.Request, key string) (int32, bool) {
	val := r.PathValue(key)
	id, err := strconv.ParseInt(val, 10, 32)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "identificador numérico inválido en la ruta")
		return 0, false
	}
	return int32(id), true
}

// itoa convierte un int32 a string.
func itoa(n int32) string {
	return strconv.FormatInt(int64(n), 10)
}

// orEmpty retorna un slice vacío no-nil si el original era nil, evitando devolver "null" en JSON.
func orEmpty[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

// dbError traduce errores de base de datos a códigos de estado HTTP apropiados:
// - pgx.ErrNoRows -> 404 Not Found
// - Código 23505 (unique_violation / duplicado de clave) -> 409 Conflict
// - Código 23503 (foreign_key_violation / clave foránea inválida) -> 422 Unprocessable Entity
// - Cualquier otro error -> 500 Internal Server Error
func dbError(w http.ResponseWriter, err error, notFoundMsg string) {
	if errors.Is(err, pgx.ErrNoRows) {
		if notFoundMsg == "" {
			notFoundMsg = "recurso no encontrado"
		}
		writeError(w, http.StatusNotFound, notFoundMsg)
		return
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			writeError(w, http.StatusConflict, "el recurso ya existe o entra en conflicto con un registro existente")
			return
		case "23503": // foreign_key_violation
			writeError(w, http.StatusUnprocessableEntity, "referencia a un recurso inexistente o inválido")
			return
		}
	}

	log.Printf("error de base de datos no controlado: %v", err)
	writeError(w, http.StatusInternalServerError, "error interno del servidor")
}

