package main

import (
	"net/http"

	"ejerEsp.com/ejerEsp/db/sqlc"
)

type correlativaInput struct {
	MateriaRequeridaID int32 `json:"materia_requerida_id"`
}

type updateCorrelativaInput struct {
	NuevaMateriaID   int32 `json:"nueva_materia_id"`
	MateriaIDMateria int32 `json:"materia_id_materia"`
}

const correlativaNotFound = "esa correlatividad no existe"

// POST /materias/{id}/correlativas
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

// GET /materias/{id}/correlativas/{requeridaId}
func (s *Server) getCorrelativa(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	requeridaID, ok := pathID(w, r, "requeridaId")
	if !ok {
		return
	}

	if _, err := s.queries.GetMateriaByID(r.Context(), id); err != nil {
		dbError(w, err, "materia no encontrada")
		return
	}

	correlativas, err := s.queries.GetCorrelativasDeMateria(r.Context(), id)
	if err != nil {
		dbError(w, err, correlativaNotFound)
		return
	}

	for _, c := range correlativas {
		if c.IDMateria == requeridaID {
			writeJSON(w, http.StatusOK, c)
			return
		}
	}

	writeError(w, http.StatusNotFound, correlativaNotFound)
}

// PUT /materias/{id}/correlativas/{requeridaId}
func (s *Server) updateCorrelativa(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	requeridaID, ok := pathID(w, r, "requeridaId")
	if !ok {
		return
	}

	var in updateCorrelativaInput
	if !decodeJSON(w, r, &in) {
		return
	}

	nuevaMateriaID := in.NuevaMateriaID
	if nuevaMateriaID <= 0 {
		nuevaMateriaID = in.MateriaIDMateria
	}

	switch {
	case nuevaMateriaID <= 0:
		writeValidation(w, map[string]string{"nueva_materia_id": "es obligatorio y debe ser un entero positivo"})
		return
	case nuevaMateriaID == requeridaID:
		writeValidation(w, map[string]string{"nueva_materia_id": "una materia no puede ser correlativa de sí misma"})
		return
	}

	// Verificar existencia de la nueva materia
	if _, err := s.queries.GetMateriaByID(r.Context(), nuevaMateriaID); err != nil {
		dbError(w, err, "la nueva materia no existe")
		return
	}

	// Verificar existencia de la correlatividad original
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

	// Evitar ciclos con la nueva materia
	correlativasInversas, err := s.queries.GetCorrelativasDeMateria(r.Context(), requeridaID)
	if err != nil {
		dbError(w, err, "")
		return
	}
	for _, c := range correlativasInversas {
		if c.IDMateria == nuevaMateriaID {
			writeError(w, http.StatusConflict, "la actualización formaría un ciclo de correlatividad")
			return
		}
	}

	actualizada, err := s.queries.UpdateCorrelatividad(r.Context(), sqlc.UpdateCorrelatividadParams{
		MateriaIDMateria:   id,
		IDMateriaRequerida: requeridaID,
		MateriaIDMateria_2: nuevaMateriaID,
	})
	if err != nil {
		dbError(w, err, correlativaNotFound)
		return
	}

	writeJSON(w, http.StatusOK, actualizada)
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

// GET /correlatividades
func (s *Server) listTodasCorrelatividades(w http.ResponseWriter, r *http.Request) {
	todas, err := s.queries.ListCorrelatividad(r.Context())
	if err != nil {
		dbError(w, err, "")
		return
	}
	writeJSON(w, http.StatusOK, orEmpty(todas))
}
