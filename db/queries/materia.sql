-- name: CreateMateria :one
INSERT INTO materia (nombre, anio, cuatrimestre)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetMateriaByID :one
SELECT * FROM materia
WHERE id_materia = $1;

-- name: ListMaterias :many
SELECT * FROM materia
ORDER BY anio, cuatrimestre;

-- name: UpdateMateria :one
UPDATE materia
SET nombre = $2, anio = $3, cuatrimestre = $4
WHERE id_materia = $1
RETURNING *;

-- name: DeleteMateria :exec
DELETE FROM materia
WHERE id_materia = $1;
