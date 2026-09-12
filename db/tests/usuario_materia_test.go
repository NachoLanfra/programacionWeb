package test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

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
		Anio:         1,
		Cuatrimestre: 1,
	})
	if err != nil {
		t.Fatalf("error preparando materia para el test: %v", err)
	}

	// Limpieza de datos creados al finalizar
	t.Cleanup(func() {
		_ = queries.DeleteUsuarioMateria(ctx, sqlc.DeleteUsuarioMateriaParams{
			UsuarioIDUsuario: usuario.IDUsuario,
			MateriaIDMateria: materia.IDMateria,
		})
		_ = queries.DeleteMateria(ctx, materia.IDMateria)
		_ = queries.DeleteUsuario(ctx, usuario.IDUsuario)
	})

	notaInicial := pgtype.Float4{Float32: 8, Valid: true}
	notaActualizada := pgtype.Float4{Float32: 9, Valid: true}

	// CREATE (SetEstadoMateria inserta el estado de cursada de un usuario en una materia)
	relacion, err := queries.SetEstadoMateria(ctx, sqlc.SetEstadoMateriaParams{ // Probamos la querie
		UsuarioIDUsuario: usuario.IDUsuario,
		MateriaIDMateria: materia.IDMateria,
		Estado:           "cursando",
		Nota:             notaInicial,
	})
	if err != nil { // Si hay error, cortamos la ejecucion del test
		t.Fatalf("error al asociar usuario con materia: %v", err)
	}

	if relacion.UsuarioIDUsuario != usuario.IDUsuario { // Si el id de usuario no coinicide
		t.Errorf("ID usuario incorrecto: esperado %d, obtenido %d", usuario.IDUsuario, relacion.UsuarioIDUsuario)
	}

	if relacion.MateriaIDMateria != materia.IDMateria { // Si el id de materia no coinicide
		t.Errorf("ID materia incorrecto: esperado %d, obtenido %d", materia.IDMateria, relacion.MateriaIDMateria)
	}

	if relacion.Estado != "cursando" { // Si el estado no coincide
		t.Errorf("estado incorrecto: esperado %q, obtenido %q", "cursando", relacion.Estado)
	}

	// GET (GetProgresoUsuarioMateria trae el progreso: nombre de la materia, estado y nota)
	relacionObtenida, err := queries.GetProgresoUsuarioMateria(ctx, sqlc.GetProgresoUsuarioMateriaParams{ // Usamos la querie
		UsuarioIDUsuario: usuario.IDUsuario,
		MateriaIDMateria: materia.IDMateria,
	})
	if err != nil { // Si hay error al usar la querie
		t.Fatalf("error al obtener el progreso de usuario_materia: %v", err)
	}

	if relacionObtenida.Estado != "cursando" { // Si lo obtenido no es "cursando"
		t.Errorf("estado obtenido incorrecto: esperado %q, obtenido %q", "cursando", relacionObtenida.Estado)
	}

	if relacionObtenida.IDMateria != materia.IDMateria { // Si la materia obtenida no coincide
		t.Errorf("ID materia obtenido incorrecto: esperado %d, obtenido %d", materia.IDMateria, relacionObtenida.IDMateria)
	}

	// UPDATE
	relacionActualizada, err := queries.UpdateUsuarioMateria(ctx, sqlc.UpdateUsuarioMateriaParams{ // Usamos la querie
		UsuarioIDUsuario: usuario.IDUsuario,
		MateriaIDMateria: materia.IDMateria,
		Estado:           "aprobado",
		Nota:             notaActualizada,
	})
	if err != nil { // Si hay error al actualizar
		t.Fatalf("error al actualizar estado de usuario_materia: %v", err)
	}

	if relacionActualizada.Estado != "aprobado" { // Si no cambio el estado
		t.Errorf("estado actualizado incorrecto: esperado %q, obtenido %q", "aprobado", relacionActualizada.Estado)
	}

	if relacionActualizada.Nota.Float32 != notaActualizada.Float32 { // Si no cambio la nota
		t.Errorf("nota actualizada incorrecta: esperado %v, obtenido %v", notaActualizada.Float32, relacionActualizada.Nota.Float32)
	}

	// LIST
	relaciones, err := queries.ListUsuarioMateria(ctx) // Usamos la querie
	if err != nil {                                    // Si falla la querie
		t.Fatalf("error al listar usuario_materia: %v", err)
	}

	encontrado := false // Variable para ver si se encontro la relacion
	for _, r := range relaciones {
		if r.UsuarioIDUsuario == usuario.IDUsuario && r.MateriaIDMateria == materia.IDMateria {
			encontrado = true
			break
		}
	}
	if !encontrado { // Si no la encontramos
		t.Error("la relacion usuario_materia creada no aparece en la lista")
	}

	// DELETE
	err = queries.DeleteUsuarioMateria(ctx, sqlc.DeleteUsuarioMateriaParams{ // Usamos la querie
		UsuarioIDUsuario: usuario.IDUsuario,
		MateriaIDMateria: materia.IDMateria,
	})
	if err != nil { // Error al eliminar
		t.Fatalf("error al eliminar relacion usuario_materia: %v", err)
	}

	// Verificación de borrado
	_, err = queries.GetProgresoUsuarioMateria(ctx, sqlc.GetProgresoUsuarioMateriaParams{
		UsuarioIDUsuario: usuario.IDUsuario,
		MateriaIDMateria: materia.IDMateria,
	})
	if err != pgx.ErrNoRows {
		t.Errorf("se esperaba pgx.ErrNoRows después de eliminar la relación, obtenido: %v", err)
	}
}
