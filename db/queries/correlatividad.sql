-- name: AddCorrelatividad :exec
INSERT INTO correlatividad_materia (materia_id_materia, id_materia_requerida)
VALUES ($1, $2);

-- name: GetCorrelativasDeMateria :many
SELECT m.*
FROM correlatividad_materia cm
JOIN materia m ON cm.id_materia_requerida = m.id_materia
WHERE cm.materia_id_materia = $1;
