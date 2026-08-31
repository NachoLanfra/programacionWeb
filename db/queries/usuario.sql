-- name: CreateUsuario :one
INSERT INTO usuario (nombre_apellido, email, telefono)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUsuarioByID :one
SELECT * FROM usuario
WHERE id_usuario = $1;

-- name: ListUsuarios :many
SELECT * FROM usuario
ORDER BY id_usuario;

-- name: UpdateUsuario :one
UPDATE usuario
SET nombre_apellido = $2,
    email = $3,
    telefono = $4
WHERE id_usuario = $1
RETURNING *;

-- name: DeleteUsuario :exec
DELETE FROM usuario
WHERE id_usuario = $1;
