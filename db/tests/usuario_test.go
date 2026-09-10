package test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"ejerEsp.com/ejerEsp/db/sqlc"
)

func TestUsuarioCRUD(t *testing.T) {
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
	usuario, err := queries.CreateUsuario(ctx, sqlc.CreateUsuarioParams{ // Probamos la querie
		NombreApellido: "Usuario Test",
		Email:          "usuario.test@example.com",
		Telefono:       "2235970000",
	})
	if err != nil { // Si hay error, cortamos al ejecucion del test
		t.Fatalf("error al crear usuario: %v", err)
	}

	if usuario.NombreApellido != "Usuario Test" { // Si el usuario no se creo correctamente, tira error
		t.Errorf(
			"nombre incorrecto: esperado %q, obtenido %q",
			"Usuario Test",
			usuario.NombreApellido,
		)
	}

	if usuario.Email != "usuario.test@example.com" { // Si el mail no se creo correctamente, tira error
		t.Errorf(
			"email incorrecto: esperado %q, obtenido %q",
			"usuario.test@example.com",
			usuario.Email,
		)
	}

	t.Cleanup(func() { // Eliminamos el usuario al finalizar el test para no dejar datos de prueba en la base
		err := queries.DeleteUsuario(ctx, usuario.IDUsuario)
		if err != nil {
			t.Errorf("error al limpiar usuario de prueba: %v", err)
		}
	})

	// GET
	usuarioObtenido, err := queries.GetUsuarioByID(ctx, usuario.IDUsuario) // Usamos la querie
	
	if err != nil {  // Si hay error al usar la querie
		t.Fatalf("error al obtener usuario: %v", err)
	}

	if usuarioObtenido.IDUsuario != usuario.IDUsuario { // Si los datos de id son incorrectos
		t.Errorf(
			"ID incorrecto: esperado %d, obtenido %d",
			usuario.IDUsuario,
			usuarioObtenido.IDUsuario,
		)
	}

	if usuarioObtenido.NombreApellido != usuario.NombreApellido { // Si los datos de nombre y apellido son incorrectos
		t.Errorf(
			"nombre incorrecto: esperado %q, obtenido %q",
			usuario.NombreApellido,
			usuarioObtenido.NombreApellido,
		)
	}

	if usuarioObtenido.Email != usuario.Email { // Si los datos de email son incorrectos
		t.Errorf(
			"email incorrecto: esperado %q, obtenido %q",
			usuario.Email,
			usuarioObtenido.Email,
		)
	}

	// LIST
	usuarios, err := queries.ListUsuarios(ctx) // Usamos la querie
	
	if err != nil { // Si falla la querie
		t.Fatalf("error al listar usuarios: %v", err)
	}

	encontrado := false // Variable para ver si se encontro el usuario

	for _, u := range usuarios { // Buscamos al usario en la lista
		if u.IDUsuario == usuario.IDUsuario {
			encontrado = true
			break
		}
	}

	if !encontrado { // Si no lo encontramos
		t.Error("el usuario creado no aparece en la lista")
	}

	// UPDATE
	usuarioActualizado, err := queries.UpdateUsuario(ctx, sqlc.UpdateUsuarioParams{ // Usamos la querie y actualizamos los datos
		IDUsuario:      usuario.IDUsuario,
		NombreApellido: "Usuario Test Actualizado",
		Email:          "usuario.actualizado@example.com",
		Telefono:       "2235971111",
	})
	
	if err != nil { // Si hay error al usar la querie
		t.Fatalf("error al actualizar usuario: %v", err)
	}

	if usuarioActualizado.NombreApellido != "Usuario Test Actualizado" { // Si el nombre y apellido no se actualiza
		t.Errorf(
			"nombre actualizado incorrecto: esperado %q, obtenido %q",
			"Usuario Test Actualizado",
			usuarioActualizado.NombreApellido,
		)
	}

	if usuarioActualizado.Email != "usuario.actualizado@example.com" { // Si el email no se actualiza
		t.Errorf(
			"email actualizado incorrecto: esperado %q, obtenido %q",
			"usuario.actualizado@example.com",
			usuarioActualizado.Email,
		)
	}

	if usuarioActualizado.Telefono != "2494111111" { // Si el numero de telefono no se actualiza
		t.Errorf(
			"teléfono actualizado incorrecto: esperado %q, obtenido %q",
			"2494111111",
			usuarioActualizado.Telefono,
		)
	}

	// DELETE
	err = queries.DeleteUsuario(ctx, usuario.IDUsuario) // Usamos la querie
	
	if err != nil { // Error al eliminar
		t.Fatalf("error al eliminar usuario: %v", err)
	}

	_, err = queries.GetUsuarioByID(ctx, usuario.IDUsuario) // Intentamos obtener el usuario eliminado para verificar que ya no exista
	if err != pgx.ErrNoRows {
		t.Errorf(
			"se esperaba pgx.ErrNoRows después de eliminar el usuario, obtenido: %v",
			err,
		)
	}
}
