package test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"ejerEsp.com/ejerEsp/db/sqlc"
)

func TestMateriaCRUD(t *testing.T) {
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

	// CREATE
	materia, err := queries.CreateMateria(ctx, sqlc.CreateMateriaParams{ // Probamos la querie
		Nombre:       "Materia Test",
		Departamento: "Computacion",
	})
	if err != nil { // Si hay error, cortamos la ejecucion del test
		t.Fatalf("error al crear materia: %v", err)
	}

	if materia.Nombre != "Materia Test" { // Si la materia no se creo correctamente, tira error
		t.Errorf(
			"nombre incorrecto: esperado %q, obtenido %q",
			"Materia Test",
			materia.Nombre,
		)
	}

	if materia.Departamento != "Computacion" { // Si el departamento no se creo correctamente, tira error
		t.Errorf(
			"departamento incorrecto: esperado %q, obtenido %q",
			"Computacion",
			materia.Departamento,
		)
	}

	t.Cleanup(func() { // Eliminamos la materia al finalizar el test para no dejar datos de prueba en la base
		err := queries.DeleteMateria(ctx, materia.IDMateria)
		if err != nil {
			t.Errorf("error al limpiar materia de prueba: %v", err)
		}
	})

	// GET
	materiaObtenida, err := queries.GetMateriaByID(ctx, materia.IDMateria) // Usamos la querie

	if err != nil { // Si hay error al usar la querie
		t.Fatalf("error al obtener materia: %v", err)
	}

	if materiaObtenida.IDMateria != materia.IDMateria { // Si los datos de id son incorrectos
		t.Errorf(
			"ID incorrecto: esperado %d, obtenido %d",
			materia.IDMateria,
			materiaObtenida.IDMateria,
		)
	}

	if materiaObtenida.Nombre != materia.Nombre { // Si los datos de nombre son incorrectos
		t.Errorf(
			"nombre incorrecto: esperado %q, obtenido %q",
			materia.Nombre,
			materiaObtenida.Nombre,
		)
	}

	if materiaObtenida.Departamento != materia.Departamento { // Si los datos de departamento son incorrectos
		t.Errorf(
			"departamento incorrecto: esperado %q, obtenido %q",
			materia.Departamento,
			materiaObtenida.Departamento,
		)
	}

	// LIST
	materias, err := queries.ListMaterias(ctx) // Usamos la querie

	if err != nil { // Si falla la querie
		t.Fatalf("error al listar materias: %v", err)
	}

	encontrado := false // Variable para ver si encontramos la materia

	for _, m := range materias { // Buscamos la materia en la lista
		if m.IDMateria == materia.IDMateria {
			encontrado = true
			break
		}
	}

	if !encontrado { // Si no la encontramos
		t.Error("la materia creada no aparece en la lista")
	}

	// UPDATE
	materiaActualizada, err := queries.UpdateMateria(ctx, sqlc.UpdateMateriaParams{ // Usamos la querie y actualizamos los datos
		IDMateria:    materia.IDMateria,
		Nombre:       "Materia Test Actualizada",
		Departamento: "Sistemas",
	})

	if err != nil { // Si hay error al usar la querie
		t.Fatalf("error al actualizar materia: %v", err)
	}

	if materiaActualizada.Nombre != "Materia Test Actualizada" { // Si el nombre no se actualiza
		t.Errorf(
			"nombre actualizado incorrecto: esperado %q, obtenido %q",
			"Materia Test Actualizada",
			materiaActualizada.Nombre,
		)
	}

	if materiaActualizada.Departamento != "Sistemas" { // Si el departamento no se actualiza
		t.Errorf(
			"departamento actualizado incorrecto: esperado %q, obtenido %q",
			"Sistemas",
			materiaActualizada.Departamento,
		)
	}

	// DELETE
	err = queries.DeleteMateria(ctx, materia.IDMateria) // Usamos la querie

	if err != nil { // Error al eliminar
		t.Fatalf("error al eliminar materia: %v", err)
	}

	_, err = queries.GetMateriaByID(ctx, materia.IDMateria) // Intentamos obtener la materia eliminada para verificar que ya no exista
	if err != pgx.ErrNoRows {
		t.Errorf(
			"se esperaba pgx.ErrNoRows despuÃ©s de eliminar la materia, obtenido: %v",
			err,
		)
	}
}
