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

-- name: ListPlanUsuario :many
-- Todas las materias del plan, con el estado y la nota del alumno (si los cargó).
SELECT m.id_materia, m.nombre, m.anio, m.cuatrimestre,
       COALESCE(um.estado, 'sin_cursar')::text AS estado,
       um.nota
FROM materia m
LEFT JOIN usuario_materia um
       ON um.materia_id_materia = m.id_materia
      AND um.usuario_id_usuario = $1
ORDER BY m.anio, m.cuatrimestre, m.nombre;

-- name: GuardarCursada :exec
-- Crea la cursada o, si ya existía, la actualiza (upsert).
INSERT INTO usuario_materia (usuario_id_usuario, materia_id_materia, estado, nota)
VALUES ($1, $2, $3, $4)
ON CONFLICT (usuario_id_usuario, materia_id_materia)
DO UPDATE SET estado = EXCLUDED.estado, nota = EXCLUDED.nota;
