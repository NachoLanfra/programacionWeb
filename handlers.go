package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"ejerEsp.com/ejerEsp/db/sqlc"
	"ejerEsp.com/ejerEsp/views"
	"github.com/jackc/pgx/v5/pgtype"
)

// Nota mínima para regular o finalizada.
const notaMinima = 4

// Estados válidos; true si llevan nota.
var estadosValidos = map[string]bool{
	"sin_cursar": false,
	"cursando":   false,
	"regular":    true,
	"finalizada": true,
}

// Estados que deben tener las correlativas según el estado elegido.
var correlativasExigidas = map[string][]string{
	"cursando":   {"regular", "finalizada"},
	"regular":    {"regular", "finalizada"},
	"finalizada": {"finalizada"},
}

// Texto de cada estado para los mensajes.
var estadoTexto = map[string]string{
	"sin_cursar": "sin cursar",
	"cursando":   "cursando",
	"regular":    "regular",
	"finalizada": "finalizada",
}

// True si el pedido viene de HTMX.
func esHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// Redirige a / con el error en la URL (sin JavaScript).
func volverConError(w http.ResponseWriter, r *http.Request, materia, msg string) {
	q := url.Values{}
	q.Set("materia", materia)
	q.Set("error", msg)
	http.Redirect(w, r, "/?"+q.Encode(), http.StatusSeeOther)
}

// Responde un error con fragmentos (HTMX) o con redirección.
func (s *Server) fallar(w http.ResponseWriter, r *http.Request, status int, materia, msg string) {
	if !esHTMX(r) {
		volverConError(w, r, materia, msg)
		return
	}
	s.responderPlan(w, r, status, views.Aviso{Materia: materia, Mensaje: msg})
}

// Devuelve la tabla y el aviso como fragmentos para HTMX.
func (s *Server) responderPlan(w http.ResponseWriter, r *http.Request, status int, aviso views.Aviso) {
	plan, err := s.queries.ListPlanUsuario(r.Context(), s.alumnoID)
	if err != nil {
		log.Printf("error listando plan: %v", err)
		http.Error(w, "error interno del servidor", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)
	if err := views.CursadaList(plan).Render(r.Context(), w); err != nil {
		log.Printf("error renderizando la tabla: %v", err)
		return
	}
	if err := views.AvisoBox(aviso, true).Render(r.Context(), w); err != nil {
		log.Printf("error renderizando el aviso: %v", err)
	}
}

// GET /
func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	alumno, err := s.queries.GetUsuarioByID(r.Context(), s.alumnoID)
	if err != nil {
		log.Printf("error obteniendo alumno: %v", err)
		http.Error(w, "error interno del servidor", http.StatusInternalServerError)
		return
	}

	plan, err := s.queries.ListPlanUsuario(r.Context(), s.alumnoID)
	if err != nil {
		log.Printf("error listando plan: %v", err)
		http.Error(w, "error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Error que llega en la URL tras un POST fallido.
	aviso := views.Aviso{
		Materia: r.URL.Query().Get("materia"),
		Mensaje: r.URL.Query().Get("error"),
	}

	if err := views.IndexPage(alumno, plan, aviso).Render(r.Context(), w); err != nil {
		log.Printf("error renderizando la página: %v", err)
	}
}

// POST /cursadas
func (s *Server) guardarCursada(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.fallar(w, r, http.StatusBadRequest, "", "El formulario llegó mal, probá de nuevo.")
		return
	}

	// Primero la materia, para nombrarla en los errores.
	materiaID, err := strconv.Atoi(r.FormValue("materia_id"))
	if err != nil || materiaID <= 0 {
		s.fallar(w, r, http.StatusBadRequest, "", "La materia no es válida.")
		return
	}
	materia, err := s.queries.GetMateriaByID(r.Context(), int32(materiaID))
	if err != nil {
		s.fallar(w, r, http.StatusNotFound, "", "La materia no existe.")
		return
	}
	nombre := materia.Nombre

	estado := r.FormValue("estado")
	llevaNota, ok := estadosValidos[estado]
	if !ok {
		s.fallar(w, r, http.StatusUnprocessableEntity, nombre, "El estado elegido no es válido.")
		return
	}

	// Reglas de la nota según el estado.
	texto := r.FormValue("nota")
	var nota pgtype.Float4 // sin Valid = NULL en la base

	if !llevaNota {
		if texto != "" {
			s.fallar(w, r, http.StatusUnprocessableEntity, nombre, "Si está sin cursar o cursando, no puede tener nota. Borrá la nota o cambiá el estado.")
			return
		}
	} else {
		if texto == "" {
			s.fallar(w, r, http.StatusUnprocessableEntity, nombre, "Para marcarla como regular o finalizada tenés que cargar la nota.")
			return
		}
		valor, err := strconv.ParseFloat(texto, 32)
		if err != nil || valor < notaMinima || valor > 10 {
			s.fallar(w, r, http.StatusUnprocessableEntity, nombre, "La nota tiene que estar entre 4 y 10.")
			return
		}
		nota = pgtype.Float4{Float32: float32(valor), Valid: true}
	}

	// Correlativas de esta materia.
	if aceptados, exige := correlativasExigidas[estado]; exige {
		faltan, err := s.queries.CorrelativasFaltantes(r.Context(), sqlc.CorrelativasFaltantesParams{
			UsuarioID:        s.alumnoID,
			MateriaID:        materia.IDMateria,
			EstadosAceptados: aceptados,
		})
		if err != nil {
			log.Printf("error consultando correlativas: %v", err)
			s.fallar(w, r, http.StatusInternalServerError, nombre, "No se pudieron revisar las correlativas, probá de nuevo.")
			return
		}
		if len(faltan) > 0 {
			requisito := "tener regulares o finalizadas"
			if estado == "finalizada" {
				requisito = "tener finalizadas"
			}
			s.fallar(w, r, http.StatusUnprocessableEntity, nombre, "Para eso necesitás "+requisito+": "+strings.Join(faltan, ", ")+".")
			return
		}
	}

	// Materias que dependen de esta.
	bloquean, err := s.dependientesQueBloquean(r.Context(), materia.IDMateria, estado)
	if err != nil {
		log.Printf("error consultando dependientes: %v", err)
		s.fallar(w, r, http.StatusInternalServerError, nombre, "No se pudieron revisar las materias que dependen de esta, probá de nuevo.")
		return
	}
	if len(bloquean) > 0 {
		s.fallar(w, r, http.StatusUnprocessableEntity, nombre, "No la podés pasar a "+estadoTexto[estado]+" porque estas materias dependen de ella: "+strings.Join(bloquean, ", ")+".")
		return
	}

	err = s.queries.GuardarCursada(r.Context(), sqlc.GuardarCursadaParams{
		UsuarioIDUsuario: s.alumnoID,
		MateriaIDMateria: materia.IDMateria,
		Estado:           estado,
		Nota:             nota,
	})
	if err != nil {
		log.Printf("error guardando cursada: %v", err)
		s.fallar(w, r, http.StatusInternalServerError, nombre, "No se pudo guardar, probá de nuevo.")
		return
	}

	// Guardado OK: fragmentos con HTMX, redirección sin HTMX.
	if esHTMX(r) {
		s.responderPlan(w, r, http.StatusOK, views.Aviso{})
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Materias cargadas que dejarían de cumplir sus correlativas si esta pasa al estado nuevo.
func (s *Server) dependientesQueBloquean(ctx context.Context, materiaID int32, estado string) ([]string, error) {
	dependientes, err := s.queries.DependientesConEstado(ctx, sqlc.DependientesConEstadoParams{
		UsuarioID: s.alumnoID,
		MateriaID: materiaID,
	})
	if err != nil {
		return nil, err
	}
	var bloquean []string
	for _, d := range dependientes {
		if !slices.Contains(correlativasExigidas[d.Estado], estado) {
			bloquean = append(bloquean, d.Nombre+" ("+estadoTexto[d.Estado]+")")
		}
	}
	return bloquean, nil
}

// DELETE /cursadas/{materiaId}
func (s *Server) borrarCursada(w http.ResponseWriter, r *http.Request) {
	materiaID, err := strconv.Atoi(r.PathValue("materiaId"))
	if err != nil || materiaID <= 0 {
		s.avisoSolo(w, r, http.StatusBadRequest, views.Aviso{Accion: "borrar", Mensaje: "La materia no es válida."})
		return
	}
	materia, err := s.queries.GetMateriaByID(r.Context(), int32(materiaID))
	if err != nil {
		s.avisoSolo(w, r, http.StatusNotFound, views.Aviso{Accion: "borrar", Mensaje: "La materia no existe."})
		return
	}
	aviso := views.Aviso{Accion: "borrar", Materia: materia.Nombre}

	// Borrar equivale a pasarla a sin cursar.
	bloquean, err := s.dependientesQueBloquean(r.Context(), materia.IDMateria, "sin_cursar")
	if err != nil {
		log.Printf("error consultando dependientes: %v", err)
		aviso.Mensaje = "No se pudieron revisar las materias que dependen de esta, probá de nuevo."
		s.responderFila(w, r, http.StatusInternalServerError, materia.IDMateria, aviso)
		return
	}
	if len(bloquean) > 0 {
		aviso.Mensaje = "No la podés borrar porque estas materias dependen de ella: " + strings.Join(bloquean, ", ") + "."
		s.responderFila(w, r, http.StatusUnprocessableEntity, materia.IDMateria, aviso)
		return
	}

	err = s.queries.DeleteUsuarioMateria(r.Context(), sqlc.DeleteUsuarioMateriaParams{
		UsuarioIDUsuario: s.alumnoID,
		MateriaIDMateria: materia.IDMateria,
	})
	if err != nil {
		log.Printf("error borrando cursada: %v", err)
		aviso.Mensaje = "No se pudo borrar, probá de nuevo."
		s.responderFila(w, r, http.StatusInternalServerError, materia.IDMateria, aviso)
		return
	}

	// Borrado OK: la fila vuelve como sin cursar y se limpia el aviso.
	s.responderFila(w, r, http.StatusOK, materia.IDMateria, views.Aviso{})
}

// Devuelve la fila de una materia y el aviso (como hx-partial).
func (s *Server) responderFila(w http.ResponseWriter, r *http.Request, status int, materiaID int32, aviso views.Aviso) {
	plan, err := s.queries.ListPlanUsuario(r.Context(), s.alumnoID)
	if err != nil {
		log.Printf("error listando plan: %v", err)
		s.avisoSolo(w, r, http.StatusInternalServerError, views.Aviso{Accion: "borrar", Mensaje: "Error interno, probá de nuevo."})
		return
	}
	i := slices.IndexFunc(plan, func(f sqlc.ListPlanUsuarioRow) bool { return f.IDMateria == materiaID })
	if i < 0 {
		s.avisoSolo(w, r, http.StatusNotFound, views.Aviso{Accion: "borrar", Mensaje: "La materia no existe."})
		return
	}
	w.WriteHeader(status)
	if err := views.FilaConAviso(plan[i], aviso).Render(r.Context(), w); err != nil {
		log.Printf("error renderizando la fila: %v", err)
	}
}

// Muestra solo el aviso, sin tocar la fila (HX-Retarget cambia el destino).
func (s *Server) avisoSolo(w http.ResponseWriter, r *http.Request, status int, aviso views.Aviso) {
	w.Header().Set("HX-Retarget", "#aviso")
	w.Header().Set("HX-Reswap", "innerHTML")
	w.WriteHeader(status)
	if err := views.AvisoContenido(aviso).Render(r.Context(), w); err != nil {
		log.Printf("error renderizando el aviso: %v", err)
	}
}
