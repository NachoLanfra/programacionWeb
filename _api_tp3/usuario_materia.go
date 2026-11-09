package main

import (
	"net/http"
	"strings"

	"ejerEsp.com/ejerEsp/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

// Estados válidos de cursada de un alumno en una materia
var estadosValidos = map[string]bool{
	"sin_cursar":  true,
	"cursando":    true,
	"aprobada":    true,
	"desaprobada": true,
}

const (
	notaMin = 0
	notaMax = 10
)

// usuarioMateriaUpdate es el JSON de PUT /usuarios/{id}/materias/{materiaId}
// La nota es opcional (null o ausente = sin nota)
type usuarioMateriaUpdate struct {
	Estado string   `json:"estado"`
	Nota   *float32 `json:"nota"`
}

// usuarioMateriaCreate es el JSON de POST /usuarios/{id}/materias
type usuarioMateriaCreate struct {
	MateriaID int32 `json:"materia_id"`
	usuarioMateriaUpdate
}

func (in *usuarioMateriaUpdate) validate() map[string]string {
	in.Estado = strings.ToLower(strings.TrimSpace(in.Estado))

	errs := map[string]string{}
	switch {
	case in.Estado == "":
		errs["estado"] = "es obligatorio"
	case !estadosValidos[in.Estado]:
		errs["estado"] = "debe ser sin_cursar, cursando, aprobada o desaprobada"
	}
	if in.Nota != nil && (*in.Nota < notaMin || *in.Nota > notaMax) {
		errs["nota"] = "debe estar entre 0 y 10"
	}
	return errs
}

func (in *usuarioMateriaCreate) validate() map[string]string {
	errs := in.usuarioMateriaUpdate.validate()
	if in.MateriaID <= 0 {
		errs["materia_id"] = "es obligatorio y debe ser un entero positivo"
	}
	return errs
}

// nota convierte el *float32 opcional del JSON al tipo nullable de pgx
func (in *usuarioMateriaUpdate) nota() pgtype.Float4 {
	if in.Nota == nil {
		return pgtype.Float4{}
	}
	return pgtype.Float4{Float32: *in.Nota, Valid: true}
}

const usuarioMateriaNotFound = "el usuario no tiene registrada esa materia"

// POST /usuarios/{id}/materias
func (s *Server) createUsuarioMateria(w http.ResponseWriter, r *http.Request) {
	usuarioID, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var in usuarioMateriaCreate
	if !decodeJSON(w, r, &in) {
		return
	}
	if errs := in.validate(); len(errs) > 0 {
		writeValidation(w, errs)
		return
	}
	if _, err := s.queries.GetUsuarioByID(r.Context(), usuarioID); err != nil {
		dbError(w, err, "usuario no encontrado")
		return
	}

	// Si la materia no existe, la FK falla y dbError responde 422
	// Si el usuario ya tiene esa materia, la PK falla y dbError responde 409
	if _, err := s.queries.SetEstadoMateria(r.Context(), sqlc.SetEstadoMateriaParams{
		UsuarioIDUsuario: usuarioID,
		MateriaIDMateria: in.MateriaID,
		Estado:           in.Estado,
		Nota:             in.nota(),
	}); err != nil {
		dbError(w, err, usuarioMateriaNotFound)
		return
	}

	p, err := s.queries.GetProgresoUsuarioMateria(r.Context(), sqlc.GetProgresoUsuarioMateriaParams{
		UsuarioIDUsuario: usuarioID,
		MateriaIDMateria: in.MateriaID,
	})
	if err != nil {
		dbError(w, err, usuarioMateriaNotFound)
		return
	}
	w.Header().Set("Location", "/usuarios/"+itoa(usuarioID)+"/materias/"+itoa(in.MateriaID))
	writeJSON(w, http.StatusCreated, p)
}

// GET /usuarios/{id}/materias
func (s *Server) listUsuarioMateria(w http.ResponseWriter, r *http.Request) {
	usuarioID, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	if _, err := s.queries.GetUsuarioByID(r.Context(), usuarioID); err != nil {
		dbError(w, err, "usuario no encontrado")
		return
	}
	all, err := s.queries.ListUsuarioMateria(r.Context())
	if err != nil {
		dbError(w, err, "")
		return
	}
	var umList []sqlc.UsuarioMaterium
	for _, item := range all {
		if item.UsuarioIDUsuario == usuarioID {
			umList = append(umList, item)
		}
	}
	writeJSON(w, http.StatusOK, orEmpty(umList))
}

// GET /usuarios/{id}/materias/{materiaId}
func (s *Server) getUsuarioMateria(w http.ResponseWriter, r *http.Request) {
	usuarioID, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	materiaID, ok := pathID(w, r, "materiaId")
	if !ok {
		return
	}
	p, err := s.queries.GetProgresoUsuarioMateria(r.Context(), sqlc.GetProgresoUsuarioMateriaParams{
		UsuarioIDUsuario: usuarioID,
		MateriaIDMateria: materiaID,
	})
	if err != nil {
		dbError(w, err, usuarioMateriaNotFound)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

// PUT /usuarios/{id}/materias/{materiaId}
func (s *Server) updateUsuarioMateria(w http.ResponseWriter, r *http.Request) {
	usuarioID, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	materiaID, ok := pathID(w, r, "materiaId")
	if !ok {
		return
	}
	var in usuarioMateriaUpdate
	if !decodeJSON(w, r, &in) {
		return
	}
	if errs := in.validate(); len(errs) > 0 {
		writeValidation(w, errs)
		return
	}

	if _, err := s.queries.UpdateUsuarioMateria(r.Context(), sqlc.UpdateUsuarioMateriaParams{
		UsuarioIDUsuario: usuarioID,
		MateriaIDMateria: materiaID,
		Estado:           in.Estado,
		Nota:             in.nota(),
	}); err != nil {
		dbError(w, err, usuarioMateriaNotFound)
		return
	}

	p, err := s.queries.GetProgresoUsuarioMateria(r.Context(), sqlc.GetProgresoUsuarioMateriaParams{
		UsuarioIDUsuario: usuarioID,
		MateriaIDMateria: materiaID,
	})
	if err != nil {
		dbError(w, err, usuarioMateriaNotFound)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

// DELETE /usuarios/{id}/materias/{materiaId}
func (s *Server) deleteUsuarioMateria(w http.ResponseWriter, r *http.Request) {
	usuarioID, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	materiaID, ok := pathID(w, r, "materiaId")
	if !ok {
		return
	}
	if _, err := s.queries.GetProgresoUsuarioMateria(r.Context(), sqlc.GetProgresoUsuarioMateriaParams{
		UsuarioIDUsuario: usuarioID,
		MateriaIDMateria: materiaID,
	}); err != nil {
		dbError(w, err, usuarioMateriaNotFound)
		return
	}
	if err := s.queries.DeleteUsuarioMateria(r.Context(), sqlc.DeleteUsuarioMateriaParams{
		UsuarioIDUsuario: usuarioID,
		MateriaIDMateria: materiaID,
	}); err != nil {
		dbError(w, err, usuarioMateriaNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
