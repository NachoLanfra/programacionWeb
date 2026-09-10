package test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"ejerEsp.com/ejerEsp/db/sqlc"
)

func TestUsuarioMateriaCRUD(t *testing.T) {
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

	// PREPARACION: Creamos un usuario y una materia necesarios para la clave foránea
	usuario, err := queries.CreateUsuario(ctx, sqlc.CreateUsuarioParams{
		NombreApellido: "Usuario Relacion Test",
		Email:          "relacion.test@example.com",
		Telefono:       "2230000000",
	})
	if err != nil {
		t.Fatalf("error preparando usuario para el test: %v", err)
	}

	materia, err := queries.CreateMateria(ctx, sqlc.CreateMateriaParams{
		Nombre:       "Materia Relacion Test",
		Departamento: "Sistemas",
	})
	if err != nil {
		t.Fatalf("error preparando materia para el test: %v", err)
	}

	// Limpieza de datos creados al finalizar
	t.Cleanup(func() {
		_ = queries.DeleteUsuarioMateria(ctx, sqlc.DeleteUsuarioMateriaParams{
			IDUsuario: usuario.IDUsuario,
			IDMateria: materia.IDMateria,
		})
		_ = queries.DeleteMateria(ctx, materia.IDMateria)
		_ = queries.DeleteUsuario(ctx, usuario.IDUsuario)
	})

	// CREATE
	relacion, err := queries.CreateUsuarioMateria(ctx, sqlc.CreateUsuarioMateriaParams{ // Probamos la querie
		IDUsuario: usuario.IDUsuario,
		IDMateria: materia.IDMateria,
		Estado:    "cursando",
	})
	if err != nil { // Si hay error, cortamos la ejecucion del test
		t.Fatalf("error al asociar usuario con materia: %v", err)
	}

	if relacion.IDUsuario != usuario.IDUsuario { // Si el id de usuario no coinicide
		t.Errorf("ID usuario incorrecto: esperado %d, obtenido %d", usuario.IDUsuario, relacion.IDUsuario)
	}

	if relacion.IDMateria != materia.IDMateria { // Si el id de materia no coinicide
		t.Errorf("ID materia incorrecto: esperado %d, obtenido %d", materia.IDMateria, relacion.IDMateria)
	}

	if relacion.Estado != "cursando" { // Si el estado no coincide
		t.Errorf("estado incorrecto: esperado %q, obtenido %q", "cursando", relacion.Estado)
	}

	// GET 
	relacionObtenida, err := queries.GetUsuarioMateria(ctx, sqlc.GetUsuarioMateriaParams{ // Usamos la querie
		IDUsuario: usuario.IDUsuario,
		IDMateria: materia.IDMateria,
	})
	if err != nil { // Si hay error al usar la querie
		t.Fatalf("error al obtener la relacion usuario_materia: %v", err)
	}

	if relacionObtenida.Estado != "cursando" { // Si lo obtenido no es "cursando"
		t.Errorf("estado obtenido incorrecto: esperado %q, obtenido %q", "cursando", relacionObtenida.Estado)
	}

	// UPDATE
	relacionActualizada, err := queries.UpdateUsuarioMateria(ctx, sqlc.UpdateUsuarioMateriaParams{ // Usamos la querie
		IDUsuario: usuario.IDUsuario,
		IDMateria: materia.IDMateria,
		Estado:    "aprobado",
	})
	if err != nil { // Si hay error al actualizar
		t.Fatalf("error al actualizar estado de usuario_materia: %v", err)
	}

	if relacionActualizada.Estado != "aprobado" { // Si no cambio el estado
		t.Errorf("estado actualizado incorrecto: esperado %q, obtenido %q", "aprobado", relacionActualizada.Estado)
	}

	// DELETE 
	err = queries.DeleteUsuarioMateria(ctx, sqlc.DeleteUsuarioMateriaParams{ // Usamos la querie
		IDUsuario: usuario.IDUsuario,
		IDMateria: materia.IDMateria,
	})
	if err != nil { // Error al eliminar
		t.Fatalf("error al eliminar relacion usuario_materia: %v", err)
	}

	// Verificación de borrado
	_, err = queries.GetUsuarioMateria(ctx, sqlc.GetUsuarioMateriaParams{
		IDUsuario: usuario.IDUsuario,
		IDMateria: materia.IDMateria,
	})
	if err != pgx.ErrNoRows {
		t.Errorf("se esperaba pgx.ErrNoRows después de eliminar la relación, obtenido: %v", err)
	}
}
