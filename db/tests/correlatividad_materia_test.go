package test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"ejerEsp.com/ejerEsp/db/sqlc"
)

func TestCorrelatividadMateriaCRUD(t *testing.T) {
	ctx := context.Background()

	conn, err := pgx.Connect(
		ctx,
		"postgres://postgres:postgres@localhost:5432/programacion_web",
	)
	if err != nil {
		t.Fatalf("no se pudo conectar a PostgreSQL: %v", err)
	}
	defer conn.Close(ctx)

	queries := sqlc.New(conn)

	// PREPARACION: Creamos las dos materias necesarias para la correlatividad
	materiaBase, err := queries.CreateMateria(ctx, sqlc.CreateMateriaParams{
		Nombre:       "Programación I",
		Departamento: "Computación",
	})
	if err != nil {
		t.Fatalf("error preparando materia base para el test: %v", err)
	}

	materiaAvanzada, err := queries.CreateMateria(ctx, sqlc.CreateMateriaParams{
		Nombre:       "Programación II",
		Departamento: "Computación",
	})
	if err != nil {
		t.Fatalf("error preparando materia avanzada para el test: %v", err)
	}

	// Limpieza de datos creados al finalizar
	t.Cleanup(func() {
		_ = queries.DeleteCorrelatividadMateria(ctx, sqlc.DeleteCorrelatividadMateriaParams{
			IDMateria:            materiaAvanzada.IDMateria,
			IDMateriaCorrelativa: materiaBase.IDMateria,
		})
		_ = queries.DeleteMateria(ctx, materiaAvanzada.IDMateria)
		_ = queries.DeleteMateria(ctx, materiaBase.IDMateria)
	})

	// CREATE
	relacion, err := queries.CreateCorrelatividadMateria(ctx, sqlc.CreateCorrelatividadMateriaParams{ // Usamos la querie
		IDMateria:            materiaAvanzada.IDMateria,
		IDMateriaCorrelativa: materiaBase.IDMateria,
	})
	if err != nil { // Si hay error, cortamos la ejecucion del test
		t.Fatalf("error al crear correlatividad entre materias: %v", err)
	}

	if relacion.IDMateria != materiaAvanzada.IDMateria { // Si la materia avanzada no coincide
		t.Errorf("ID materia incorrecto: esperado %d, obtenido %d", materiaAvanzada.IDMateria, relacion.IDMateria)
	}

	if relacion.IDMateriaCorrelativa != materiaBase.IDMateria { // Si la materia base no coincide
		t.Errorf("ID materia correlativa incorrecto: esperado %d, obtenido %d", materiaBase.IDMateria, relacion.IDMateriaCorrelativa)
	}

	// GET 
	relacionObtenida, err := queries.GetCorrelatividadMateria(ctx, sqlc.GetCorrelatividadMateriaParams{ // Usamos la querie
		IDMateria:            materiaAvanzada.IDMateria,
		IDMateriaCorrelativa: materiaBase.IDMateria,
	})
	if err != nil { // Si hay error al usar la querie
		t.Fatalf("error al obtener la correlatividad: %v", err)
	}

	if relacionObtenida.IDMateria != materiaAvanzada.IDMateria { // Si la materia avanzada no es la misma que creamos
		t.Errorf("ID materia obtenido incorrecto: esperado %d, obtenido %d", materiaAvanzada.IDMateria, relacionObtenida.IDMateria)
	}

	if relacionObtenida.IDMateriaCorrelativa != materiaBase.IDMateria { // Si la materia base no es la misma que creamos
		t.Errorf("ID materia obtenido incorrecto: esperado %d, obtenido %d", materiaBase.IDMateria, relacionObtenida.IDMateriaCorrelativa)
	}

	// LIST 
	correlativas, err := queries.ListCorrelatividadesByMateria(ctx, materiaAvanzada.IDMateria)
	if err != nil { // Si falla la querie
		t.Fatalf("error al listar correlatividades de la materia: %v", err)
	}

	encontrado := false // Variable para buscar la correlativa en la lista

	for _, c := range correlativas { // Buscamos la correlativa en la lista
		if c.IDMateriaCorrelativa == materiaBase.IDMateria {
			encontrado = true
			break
		}
	}

	if !encontrado { // Si no la encontramos
		t.Error("la materia correlativa creada no aparece en la lista")
	}

	// DELETE 
	err = queries.DeleteCorrelatividadMateria(ctx, sqlc.DeleteCorrelatividadMateriaParams{ // Usamos la querie
		IDMateria:            materiaAvanzada.IDMateria,
		IDMateriaCorrelativa: materiaBase.IDMateria,
	})
	if err != nil { // Error al eliminar
		t.Fatalf("error al eliminar correlatividad: %v", err)
	}

	// Verificación de borrado
	_, err = queries.GetCorrelatividadMateria(ctx, sqlc.GetCorrelatividadMateriaParams{
		IDMateria:            materiaAvanzada.IDMateria,
		IDMateriaCorrelativa: materiaBase.IDMateria,
	})
	if err != pgx.ErrNoRows {
		t.Errorf("se esperaba pgx.ErrNoRows después de eliminar la correlatividad, obtenido: %v", err)
	}
}
