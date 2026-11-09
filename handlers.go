package main

import (
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

// notaMinima es la nota mínima para regularizar o finalizar una materia.
const notaMinima = 4

// estadosValidos son los únicos valores que acepta una cursada.
var estadosValidos = map[string]bool{
	"sin_cursar": false,
	"cursando":   false,
	"regular":    true,
	"finalizada": true,
}

// correlativasExigidas indica, para cada estado, en qué estados tienen que
// estar TODAS sus correlativas. Sin cursar no exige nada.
var correlativasExigidas = map[string][]string{
	"cursando":   {"regular", "finalizada"},
	"regular":    {"regular", "finalizada"},
	"finalizada": {"finalizada"},
}

// estadoTexto es cómo se muestra cada estado en los mensajes.
var estadoTexto = map[string]string{
	"sin_cursar": "sin cursar",
	"cursando":   "cursando",
	"regular":    "regular",
	"finalizada": "finalizada",
}

// volverConError redirige a la página principal llevando la materia y el
// mensaje en la URL. Sigue siendo Post/Redirect/Get: termina en un GET a /.
func volverConError(w http.ResponseWriter, r *http.Request, materia, msg string) {
	q := url.Values{}
	q.Set("materia", materia)
	q.Set("error", msg)
	http.Redirect(w, r, "/?"+q.Encode(), http.StatusSeeOther)
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

	// Si un POST falló, la materia y el mensaje llegan en la URL.
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
		volverConError(w, r, "", "El formulario llegó mal, probá de nuevo.")
		return
	}

	// Primero la materia, así todos los mensajes pueden decir cuál falló.
	materiaID, err := strconv.Atoi(r.FormValue("materia_id"))
	if err != nil || materiaID <= 0 {
		volverConError(w, r, "", "La materia no es válida.")
		return
	}
	materia, err := s.queries.GetMateriaByID(r.Context(), int32(materiaID))
	if err != nil {
		volverConError(w, r, "", "La materia no existe.")
		return
	}
	nombre := materia.Nombre

	estado := r.FormValue("estado")
	llevaNota, ok := estadosValidos[estado]
	if !ok {
		volverConError(w, r, nombre, "El estado elegido no es válido.")
		return
	}

	// Reglas de la nota según el estado.
	texto := r.FormValue("nota")
	var nota pgtype.Float4 // sin Valid = NULL en la base

	if !llevaNota {
		if texto != "" {
			volverConError(w, r, nombre, "Si está sin cursar o cursando, no puede tener nota. Borrá la nota o cambiá el estado.")
			return
		}
	} else {
		if texto == "" {
			volverConError(w, r, nombre, "Para marcarla como regular o finalizada tenés que cargar la nota.")
			return
		}
		valor, err := strconv.ParseFloat(texto, 32)
		if err != nil || valor < notaMinima || valor > 10 {
			volverConError(w, r, nombre, "La nota tiene que estar entre 4 y 10.")
			return
		}
		nota = pgtype.Float4{Float32: float32(valor), Valid: true}
	}

	// Correlatividades: las requeridas tienen que estar en el estado que exige.
	if aceptados, exige := correlativasExigidas[estado]; exige {
		faltan, err := s.queries.CorrelativasFaltantes(r.Context(), sqlc.CorrelativasFaltantesParams{
			UsuarioID:        s.alumnoID,
			MateriaID:        materia.IDMateria,
			EstadosAceptados: aceptados,
		})
		if err != nil {
			log.Printf("error consultando correlativas: %v", err)
			volverConError(w, r, nombre, "No se pudieron revisar las correlativas, probá de nuevo.")
			return
		}
		if len(faltan) > 0 {
			requisito := "tener regulares o finalizadas"
			if estado == "finalizada" {
				requisito = "tener finalizadas"
			}
			volverConError(w, r, nombre, "Para eso necesitás "+requisito+": "+strings.Join(faltan, ", ")+".")
			return
		}
	}

	// Chequeo al revés: las materias que dependen de esta y ya están cargadas
	// tienen que seguir cumpliendo sus correlativas con el estado nuevo.
	dependientes, err := s.queries.DependientesConEstado(r.Context(), sqlc.DependientesConEstadoParams{
		UsuarioID: s.alumnoID,
		MateriaID: materia.IDMateria,
	})
	if err != nil {
		log.Printf("error consultando dependientes: %v", err)
		volverConError(w, r, nombre, "No se pudieron revisar las materias que dependen de esta, probá de nuevo.")
		return
	}
	var bloquean []string
	for _, d := range dependientes {
		if !slices.Contains(correlativasExigidas[d.Estado], estado) {
			bloquean = append(bloquean, d.Nombre+" ("+estadoTexto[d.Estado]+")")
		}
	}
	if len(bloquean) > 0 {
		volverConError(w, r, nombre, "No la podés pasar a "+estadoTexto[estado]+" porque estas materias dependen de ella: "+strings.Join(bloquean, ", ")+".")
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
		volverConError(w, r, nombre, "No se pudo guardar, probá de nuevo.")
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
