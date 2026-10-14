package main

import (
	"net/http"

	"ejerEsp.com/ejerEsp/db/sqlc"
)

// POST /materias
func (s *Server) crearMateria(w http.ResponseWriter, r *http.Request) {
	var params sqlc.CreateMateriaParams
	if !decodeJSON(w, r, &params) {
		return
	}

	if params.Nombre == "" || params.Anio <= 0 || params.Cuatrimestre <= 0 {
		writeValidation(w, map[string]string{
			"campos": "nombre, anio y cuatrimestre son obligatorios y deben ser enteros positivos",
		})
		return
	}

	materia, err := s.queries.CreateMateria(r.Context(), params)
	if err != nil {
		dbError(w, err, "")
		return
	}

	w.Header().Set("Location", "/materias/"+itoa(materia.IDMateria))
	writeJSON(w, http.StatusCreated, materia)
}

// GET /materias
func (s *Server) listarMaterias(w http.ResponseWriter, r *http.Request) {
	materias, err := s.queries.ListMaterias(r.Context())
	if err != nil {
		dbError(w, err, "")
		return
	}
	writeJSON(w, http.StatusOK, orEmpty(materias))
}

// GET /materias/{id}
func (s *Server) obtenerMateria(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}

	materia, err := s.queries.GetMateriaByID(r.Context(), id)
	if err != nil {
		dbError(w, err, "materia no encontrada")
		return
	}

	writeJSON(w, http.StatusOK, materia)
}

// PUT /materias/{id}
func (s *Server) actualizarMateria(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}

	var req struct {
		Nombre       string `json:"nombre"`
		Anio         int32  `json:"anio"`
		Cuatrimestre int32  `json:"cuatrimestre"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	if req.Nombre == "" || req.Anio <= 0 || req.Cuatrimestre <= 0 {
		writeValidation(w, map[string]string{
			"campos": "nombre, anio y cuatrimestre son obligatorios y deben ser enteros positivos",
		})
		return
	}

	materia, err := s.queries.UpdateMateria(r.Context(), sqlc.UpdateMateriaParams{
		IDMateria:    id,
		Nombre:       req.Nombre,
		Anio:         req.Anio,
		Cuatrimestre: req.Cuatrimestre,
	})
	if err != nil {
		dbError(w, err, "materia no encontrada")
		return
	}

	writeJSON(w, http.StatusOK, materia)
}

// DELETE /materias/{id}
func (s *Server) eliminarMateria(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}

	if _, err := s.queries.GetMateriaByID(r.Context(), id); err != nil {
		dbError(w, err, "materia no encontrada")
		return
	}

	if err := s.queries.DeleteMateria(r.Context(), id); err != nil {
		dbError(w, err, "materia no encontrada")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
