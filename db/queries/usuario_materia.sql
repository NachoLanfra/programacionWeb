-- name: SetEstadoMateria :exec
INSERT INTO usuario_materia (usuario_id_usuario, materia_id_materia, estado, nota)
VALUES ($1, $2, $3, $4)
ON CONFLICT (usuario_id_usuario, materia_id_materia)
DO UPDATE SET estado = EXCLUDED.estado, nota = EXCLUDED.nota;

-- name: GetProgresoUsuario :many
SELECT m.id_materia, m.nombre, um.estado, um.nota
FROM usuario_materia um
JOIN materia m ON um.materia_id_materia = m.id_materia
WHERE um.usuario_id_usuario = $1;
