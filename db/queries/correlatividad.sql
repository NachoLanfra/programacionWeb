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

-- name: CorrelativasFaltantes :many
-- Nombres de las correlativas de una materia cuyo estado (para ese alumno)
-- NO está entre los estados aceptados.
SELECT r.nombre
FROM correlatividad_materia c
JOIN materia r ON r.id_materia = c.id_materia_requerida
LEFT JOIN usuario_materia um
       ON um.materia_id_materia = r.id_materia
      AND um.usuario_id_usuario = sqlc.arg(usuario_id)
WHERE c.materia_id_materia = sqlc.arg(materia_id)
  AND NOT (COALESCE(um.estado, 'sin_cursar') = ANY(sqlc.arg(estados_aceptados)::text[]))
ORDER BY r.anio, r.cuatrimestre, r.nombre;

-- name: DependientesConEstado :many
-- Materias que tienen a esta como correlativa y que el alumno ya cargó
-- (con un estado distinto de sin cursar), junto con ese estado.
SELECT m.nombre, um.estado
FROM correlatividad_materia c
JOIN materia m ON m.id_materia = c.materia_id_materia
JOIN usuario_materia um
     ON um.materia_id_materia = m.id_materia
    AND um.usuario_id_usuario = sqlc.arg(usuario_id)
WHERE c.id_materia_requerida = sqlc.arg(materia_id)
  AND um.estado <> 'sin_cursar'
ORDER BY m.anio, m.cuatrimestre, m.nombre;
