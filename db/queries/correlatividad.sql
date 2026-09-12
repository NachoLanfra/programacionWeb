-- name: AddCorrelatividad :one
INSERT INTO correlatividad_materia (materia_id_materia, id_materia_requerida)
VALUES ($1, $2)
RETURNING *;

-- name: GetCorrelativasDeMateria :many
SELECT m.*
FROM correlatividad_materia cm
JOIN materia m ON cm.id_materia_requerida = m.id_materia
WHERE cm.materia_id_materia = $1;

-- name: ListCorrelatividad :many
SELECT *
FROM correlatividad_materia;

-- name: UpdateCorrelatividad :one
UPDATE correlatividad_materia
SET materia_id_materia = $3
WHERE materia_id_materia = $1 AND id_materia_requerida = $2
RETURNING *;

-- name: DeleteCorrelatividad :exec
DELETE FROM correlatividad_materia
WHERE materia_id_materia = $1 AND id_materia_requerida = $2;
