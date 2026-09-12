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

	// PREPARACION).
	materiaBase, err := queries.CreateMateria(ctx, sqlc.CreateMateriaParams{
		Nombre:       "Análisis Matemático I",
		Anio:         1,
		Cuatrimestre: 1,
	})
	if err != nil {
		t.Fatalf("error preparando materia base para el test: %v", err)
	}

	materiaAvanzada1, err := queries.CreateMateria(ctx, sqlc.CreateMateriaParams{
		Nombre:       "Análisis Matemático II",
		Anio:         1,
		Cuatrimestre: 2,
	})
	if err != nil {
		t.Fatalf("error preparando materia avanzada 1 para el test: %v", err)
	}

	materiaAvanzada2, err := queries.CreateMateria(ctx, sqlc.CreateMateriaParams{
		Nombre:       "Probabilidad y Estadística",
		Anio:         2,
		Cuatrimestre: 1,
	})
	if err != nil {
		t.Fatalf("error preparando materia avanzada 2 para el test: %v", err)
	}

	// Limpieza de datos creados al finalizar.
	t.Cleanup(func() {
		_ = queries.DeleteCorrelatividad(ctx, sqlc.DeleteCorrelatividadParams{
			MateriaIDMateria:   materiaAvanzada2.IDMateria,
			IDMateriaRequerida: materiaBase.IDMateria,
		})
		_ = queries.DeleteMateria(ctx, materiaAvanzada2.IDMateria)
		_ = queries.DeleteMateria(ctx, materiaAvanzada1.IDMateria)
		_ = queries.DeleteMateria(ctx, materiaBase.IDMateria)
	})

	// CREATE (AddCorrelatividad indica que materiaAvanzada1 requiere materiaBase)
	relacion, err := queries.AddCorrelatividad(ctx, sqlc.AddCorrelatividadParams{ // Usamos la querie
		MateriaIDMateria:   materiaAvanzada1.IDMateria,
		IDMateriaRequerida: materiaBase.IDMateria,
	})
	if err != nil { // Si hay error, cortamos la ejecucion del test
		t.Fatalf("error al crear correlatividad entre materias: %v", err)
	}

	if relacion.MateriaIDMateria != materiaAvanzada1.IDMateria { // Si la materia avanzada no coincide
		t.Errorf("ID materia incorrecto: esperado %d, obtenido %d", materiaAvanzada1.IDMateria, relacion.MateriaIDMateria)
	}

	if relacion.IDMateriaRequerida != materiaBase.IDMateria { // Si la materia requerida no coincide
		t.Errorf("ID materia requerida incorrecto: esperado %d, obtenido %d", materiaBase.IDMateria, relacion.IDMateriaRequerida)
	}

	// GET 
	correlativas, err := queries.GetCorrelativasDeMateria(ctx, materiaAvanzada1.IDMateria) // Usamos la querie
	if err != nil {                                                                        // Si hay error al usar la querie
		t.Fatalf("error al obtener las correlativas de la materia: %v", err)
	}

	encontradaEnCorrelativas := false
	for _, c := range correlativas {
		if c.IDMateria == materiaBase.IDMateria {
			encontradaEnCorrelativas = true
			break
		}
	}
	if !encontradaEnCorrelativas { // Si materiaBase no aparece como correlativa de materiaAvanzada1
		t.Error("la materia base no aparece entre las correlativas de la materia avanzada")
	}

	// LIST
	todas, err := queries.ListCorrelatividad(ctx) // Usamos la querie
	if err != nil {                               // Si falla la querie
		t.Fatalf("error al listar correlatividades: %v", err)
	}

	encontrado := false // Variable para buscar la correlativa en la lista completa
	for _, c := range todas {
		if c.MateriaIDMateria == materiaAvanzada1.IDMateria && c.IDMateriaRequerida == materiaBase.IDMateria {
			encontrado = true
			break
		}
	}
	if !encontrado { // Si no la encontramos
		t.Error("la correlatividad creada no aparece en la lista completa")
	}

	// UPDATE
	actualizada, err := queries.UpdateCorrelatividad(ctx, sqlc.UpdateCorrelatividadParams{
		MateriaIDMateria:   materiaAvanzada1.IDMateria, // valor viejo, usado en el WHERE
		IDMateriaRequerida: materiaBase.IDMateria,      // no cambia, usado en el WHERE
		MateriaIDMateria_2: materiaAvanzada2.IDMateria, // valor nuevo, usado en el SET
	})
	if err != nil {
		t.Fatalf("error al actualizar la correlatividad: %v", err)
	}

	if actualizada.MateriaIDMateria != materiaAvanzada2.IDMateria { // Si no quedó asociada a la nueva materia
		t.Errorf("MateriaIDMateria tras update = %d, esperado %d", actualizada.MateriaIDMateria, materiaAvanzada2.IDMateria)
	}

	if actualizada.IDMateriaRequerida != materiaBase.IDMateria { // La materia requerida no debería haber cambiado
		t.Errorf("IDMateriaRequerida tras update = %d, esperado %d", actualizada.IDMateriaRequerida, materiaBase.IDMateria)
	}

	// La correlatividad vieja ya no debería existir
	correlativasDeAvanzada1, err := queries.GetCorrelativasDeMateria(ctx, materiaAvanzada1.IDMateria)
	if err != nil {
		t.Fatalf("error al verificar correlativas de materiaAvanzada1 tras el update: %v", err)
	}
	for _, c := range correlativasDeAvanzada1 {
		if c.IDMateria == materiaBase.IDMateria {
			t.Error("la correlatividad vieja (materiaAvanzada1 -> materiaBase) seguía existiendo tras el update")
		}
	}

	// La correlatividad nueva (materiaAvanzada2 -> materiaBase) debería existir
	correlativasDeAvanzada2, err := queries.GetCorrelativasDeMateria(ctx, materiaAvanzada2.IDMateria)
	if err != nil {
		t.Fatalf("error al verificar correlativas de materiaAvanzada2 tras el update: %v", err)
	}
	encontradaTrasUpdate := false
	for _, c := range correlativasDeAvanzada2 {
		if c.IDMateria == materiaBase.IDMateria {
			encontradaTrasUpdate = true
			break
		}
	}
	if !encontradaTrasUpdate {
		t.Error("la correlatividad nueva (materiaAvanzada2 -> materiaBase) no aparece tras el update")
	}

	// DELETE 
	err = queries.DeleteCorrelatividad(ctx, sqlc.DeleteCorrelatividadParams{ // Usamos la querie
		MateriaIDMateria:   materiaAvanzada2.IDMateria,
		IDMateriaRequerida: materiaBase.IDMateria,
	})
	if err != nil { // Error al eliminar
		t.Fatalf("error al eliminar correlatividad: %v", err)
	}

	// Verificación de borrado
	correlativasTrasBorrado, err := queries.GetCorrelativasDeMateria(ctx, materiaAvanzada2.IDMateria)
	if err != nil {
		t.Fatalf("error al verificar correlativas tras el borrado: %v", err)
	}
	for _, c := range correlativasTrasBorrado {
		if c.IDMateria == materiaBase.IDMateria {
			t.Error("la correlatividad seguía existiendo después de eliminarla")
		}
	}
}
