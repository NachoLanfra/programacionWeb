-- name: SetEstadoMateria :one
INSERT INTO usuario_materia (usuario_id_usuario, materia_id_materia, estado, nota)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetProgresoUsuarioMateria :one
SELECT m.id_materia, m.nombre, um.estado, um.nota
FROM usuario_materia um
JOIN materia m ON um.materia_id_materia = m.id_materia
WHERE um.usuario_id_usuario = $1 AND um.materia_id_materia = $2;

-- name: ListUsuarioMateria :many
SELECT * FROM usuario_materia;

-- name: UpdateUsuarioMateria :one
UPDATE usuario_materia
SET estado = $3, nota = $4
WHERE usuario_id_usuario = $1 AND materia_id_materia = $2
RETURNING *;

-- name: DeleteUsuarioMateria :exec
DELETE FROM usuario_materia
WHERE usuario_id_usuario = $1 AND materia_id_materia = $2;
