package main

import (
	"net/http"
	"net/mail"
	"strings"
	"unicode/utf8"

	"ejerEsp.com/ejerEsp/db/sqlc"
)

// usuarioInput es el JSON que reciben POST /usuarios y PUT /usuarios/{id}.
type usuarioInput struct {
	NombreApellido string `json:"nombre_apellido"`
	Email          string `json:"email"`
	Telefono       string `json:"telefono"`
}

// validate normaliza (recorta espacios) y devuelve los errores por campo.
// Los límites de longitud coinciden con las columnas VARCHAR del esquema.
func (in *usuarioInput) validate() map[string]string {
	in.NombreApellido = strings.TrimSpace(in.NombreApellido)
	in.Email = strings.TrimSpace(in.Email)
	in.Telefono = strings.TrimSpace(in.Telefono)

	errs := map[string]string{}

	switch {
	case in.NombreApellido == "":
		errs["nombre_apellido"] = "es obligatorio"
	case utf8.RuneCountInString(in.NombreApellido) > 100:
		errs["nombre_apellido"] = "máximo 100 caracteres"
	}

	switch {
	case in.Email == "":
		errs["email"] = "es obligatorio"
	case utf8.RuneCountInString(in.Email) > 100:
		errs["email"] = "máximo 100 caracteres"
	case !validEmail(in.Email):
		errs["email"] = "no tiene un formato válido"
	}

	switch {
	case in.Telefono == "":
		errs["telefono"] = "es obligatorio"
	case utf8.RuneCountInString(in.Telefono) > 20:
		errs["telefono"] = "máximo 20 caracteres"
	}

	return errs
}

// validEmail acepta solo direcciones simples "local@dominio" (sin nombre
// visible tipo "Juan <juan@x.com>").
func validEmail(s string) bool {
	addr, err := mail.ParseAddress(s)
	return err == nil && addr.Address == s && strings.Contains(s[strings.LastIndex(s, "@"):], ".")
}

// POST /usuarios
func (s *Server) createUsuario(w http.ResponseWriter, r *http.Request) {
	var in usuarioInput
	if !decodeJSON(w, r, &in) {
		return
	}
	if errs := in.validate(); len(errs) > 0 {
		writeValidation(w, errs)
		return
	}

	u, err := s.queries.CreateUsuario(r.Context(), sqlc.CreateUsuarioParams{
		NombreApellido: in.NombreApellido,
		Email:          in.Email,
		Telefono:       in.Telefono,
	})
	if err != nil {
		dbError(w, err, "usuario no encontrado")
		return
	}
	w.Header().Set("Location", "/usuarios/"+itoa(u.IDUsuario))
	writeJSON(w, http.StatusCreated, u)
}

// GET /usuarios
func (s *Server) listUsuarios(w http.ResponseWriter, r *http.Request) {
	us, err := s.queries.ListUsuarios(r.Context())
	if err != nil {
		dbError(w, err, "")
		return
	}
	writeJSON(w, http.StatusOK, orEmpty(us))
}

// GET /usuarios/{id}
func (s *Server) getUsuario(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	u, err := s.queries.GetUsuarioByID(r.Context(), id)
	if err != nil {
		dbError(w, err, "usuario no encontrado")
		return
	}
	writeJSON(w, http.StatusOK, u)
}

// PUT /usuarios/{id}
func (s *Server) updateUsuario(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var in usuarioInput
	if !decodeJSON(w, r, &in) {
		return
	}
	if errs := in.validate(); len(errs) > 0 {
		writeValidation(w, errs)
		return
	}

	u, err := s.queries.UpdateUsuario(r.Context(), sqlc.UpdateUsuarioParams{
		IDUsuario:      id,
		NombreApellido: in.NombreApellido,
		Email:          in.Email,
		Telefono:       in.Telefono,
	})
	if err != nil {
		dbError(w, err, "usuario no encontrado")
		return
	}
	writeJSON(w, http.StatusOK, u)
}

// DELETE /usuarios/{id}
// DeleteUsuario no falla si el id no existe, así que primero lo buscamos
// para poder responder 404 en ese caso.
func (s *Server) deleteUsuario(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	if _, err := s.queries.GetUsuarioByID(r.Context(), id); err != nil {
		dbError(w, err, "usuario no encontrado")
		return
	}
	if err := s.queries.DeleteUsuario(r.Context(), id); err != nil {
		dbError(w, err, "usuario no encontrado")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
