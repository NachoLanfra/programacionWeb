package main

import (
	"net/http"

	"ejerEsp.com/ejerEsp/db/sqlc"
)

// correlativaInput es el JSON de POST /materias/{id}/correlativas:
// la materia {id} pasa a requerir tener aprobada "materia_requerida_id".
type correlativaInput struct {
	MateriaRequeridaID int32 `json:"materia_requerida_id"`
}

const correlativaNotFound = "esa correlatividad no existe"

// POST /materias/{id}/correlativas
//
// La relación no tiene atributos que se puedan modificar (es solo un par de
// materias), por eso no hay PUT: para "cambiarla" se borra y se vuelve a crear.
func (s *Server) createCorrelativa(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var in correlativaInput
	if !decodeJSON(w, r, &in) {
		return
	}
	switch {
	case in.MateriaRequeridaID <= 0:
		writeValidation(w, map[string]string{"materia_requerida_id": "es obligatorio y debe ser un entero positivo"})
		return
	case in.MateriaRequeridaID == id:
		writeValidation(w, map[string]string{"materia_requerida_id": "una materia no puede ser correlativa de sí misma"})
		return
	}

	if _, err := s.queries.GetMateriaByID(r.Context(), id); err != nil {
		dbError(w, err, "materia no encontrada")
		return
	}

	// Evita el ciclo más simple (A requiere B y B requiere A).
	correlativasInversas, err := s.queries.GetCorrelativasDeMateria(r.Context(), in.MateriaRequeridaID)
	if err != nil {
		dbError(w, err, "")
		return
	}
	for _, c := range correlativasInversas {
		if c.IDMateria == id {
			writeError(w, http.StatusConflict, "ya existe la correlatividad inversa; se formaría un ciclo")
			return
		}
	}

	// Si la materia requerida no existe, la FK falla y dbError responde 422.
	// Si la correlatividad ya existe, la PK falla y dbError responde 409.
	c, err := s.queries.AddCorrelatividad(r.Context(), sqlc.AddCorrelatividadParams{
		MateriaIDMateria:   id,
		IDMateriaRequerida: in.MateriaRequeridaID,
	})
	if err != nil {
		dbError(w, err, correlativaNotFound)
		return
	}
	w.Header().Set("Location", "/materias/"+itoa(id)+"/correlativas")
	writeJSON(w, http.StatusCreated, c)
}

// GET /materias/{id}/correlativas
// Devuelve las materias que hay que tener aprobadas para cursar la materia {id}.
func (s *Server) listCorrelativas(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	if _, err := s.queries.GetMateriaByID(r.Context(), id); err != nil {
		dbError(w, err, "materia no encontrada")
		return
	}
	ms, err := s.queries.GetCorrelativasDeMateria(r.Context(), id)
	if err != nil {
		dbError(w, err, "")
		return
	}
	writeJSON(w, http.StatusOK, orEmpty(ms))
}

// DELETE /materias/{id}/correlativas/{requeridaId}
func (s *Server) deleteCorrelativa(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	requeridaID, ok := pathID(w, r, "requeridaId")
	if !ok {
		return
	}

	correlativas, err := s.queries.GetCorrelativasDeMateria(r.Context(), id)
	if err != nil {
		dbError(w, err, correlativaNotFound)
		return
	}
	encontrada := false
	for _, c := range correlativas {
		if c.IDMateria == requeridaID {
			encontrada = true
			break
		}
	}
	if !encontrada {
		writeError(w, http.StatusNotFound, correlativaNotFound)
		return
	}

	if err := s.queries.DeleteCorrelatividad(r.Context(), sqlc.DeleteCorrelatividadParams{
		MateriaIDMateria:   id,
		IDMateriaRequerida: requeridaID,
	}); err != nil {
		dbError(w, err, correlativaNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
