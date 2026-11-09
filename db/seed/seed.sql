-- Plan de estudios 2023 de Ingeniería de Sistemas (UNICEN).
-- Es idempotente: se puede correr varias veces sin duplicar datos.

INSERT INTO materia (nombre, anio, cuatrimestre)
VALUES
    -- 1° año
    ('Introducción a la Programación I', 1, 1),
    ('Introducción al Álgebra', 1, 1),
    ('Introducción a los Sistemas Informáticos', 1, 1),
    ('Inglés', 1, 1),
    ('Diseño Lógico', 1, 2),
    ('Introducción a la Programación II', 1, 2),
    ('Lenguajes Formales y Autómatas', 1, 2),
    ('Cálculo I', 1, 2),
    -- 2° año
    ('Arquitectura de Computadoras I', 2, 1),
    ('Análisis y Diseño de Algoritmos I', 2, 1),
    ('Programación Orientada a Objetos', 2, 1),
    ('Álgebra Lineal', 2, 1),
    ('Redes de Computadoras I', 2, 2),
    ('Metodologías de Desarrollo de Software', 2, 2),
    ('Análisis y Diseño de Algoritmos II', 2, 2),
    ('Cálculo II', 2, 2),
    -- 3° año
    ('Sistemas Operativos', 3, 1),
    ('Base de Datos I', 3, 1),
    ('Física I', 3, 1),
    ('Lenguajes y Paradigmas', 3, 1),
    ('Programación WEB', 3, 2),
    ('Base de Datos II', 3, 2),
    ('Probabilidad y Estadística', 3, 2),
    ('Ingeniería de Software I', 3, 2),
    -- 4° año
    ('Organización y Gestión Empresarial', 4, 1),
    ('Teoría de la Información', 4, 1),
    ('Redes de Computadoras II', 4, 1),
    ('Física II', 4, 1),
    ('Calidad de Software', 4, 2),
    ('Arquitectura de Computadoras II', 4, 2),
    ('Fundamentos de la Ciencia de Datos', 4, 2),
    ('Compiladores e Intérpretes', 4, 2),
    -- 5° año
    ('Ingeniería de Software II', 5, 1),
    ('Ciberseguridad', 5, 1),
    ('Inteligencia Artificial', 5, 1),
    ('Formulación y Evaluación de Proyectos TICs', 5, 2),
    ('Ética y Legislación de la Práctica Profesional', 5, 2),
    ('Práctica Profesional Supervisada y Proyecto Integrador', 5, 2)
ON CONFLICT (nombre) DO NOTHING;

-- Correlatividades: (materia, materia requerida), buscadas por nombre.
INSERT INTO correlatividad_materia (materia_id_materia, id_materia_requerida)
SELECT m.id_materia, r.id_materia
FROM (VALUES
    -- 1° año
    ('Diseño Lógico', 'Introducción a los Sistemas Informáticos'),
    ('Diseño Lógico', 'Introducción al Álgebra'),
    ('Introducción a la Programación II', 'Introducción a la Programación I'),
    ('Lenguajes Formales y Autómatas', 'Introducción al Álgebra'),
    -- 2° año
    ('Arquitectura de Computadoras I', 'Diseño Lógico'),
    ('Análisis y Diseño de Algoritmos I', 'Introducción a la Programación II'),
    ('Análisis y Diseño de Algoritmos I', 'Cálculo I'),
    ('Análisis y Diseño de Algoritmos I', 'Introducción al Álgebra'),
    ('Programación Orientada a Objetos', 'Introducción a la Programación II'),
    ('Álgebra Lineal', 'Introducción al Álgebra'),
    ('Redes de Computadoras I', 'Introducción a la Programación II'),
    ('Redes de Computadoras I', 'Introducción a los Sistemas Informáticos'),
    ('Metodologías de Desarrollo de Software', 'Programación Orientada a Objetos'),
    ('Análisis y Diseño de Algoritmos II', 'Análisis y Diseño de Algoritmos I'),
    ('Cálculo II', 'Cálculo I'),
    -- 3° año
    ('Sistemas Operativos', 'Arquitectura de Computadoras I'),
    ('Sistemas Operativos', 'Análisis y Diseño de Algoritmos II'),
    ('Base de Datos I', 'Análisis y Diseño de Algoritmos I'),
    ('Física I', 'Álgebra Lineal'),
    ('Lenguajes y Paradigmas', 'Arquitectura de Computadoras I'),
    ('Programación WEB', 'Sistemas Operativos'),
    ('Programación WEB', 'Base de Datos I'),
    ('Programación WEB', 'Metodologías de Desarrollo de Software'),
    ('Base de Datos II', 'Base de Datos I'),
    ('Probabilidad y Estadística', 'Cálculo I'),
    ('Ingeniería de Software I', 'Sistemas Operativos'),
    ('Ingeniería de Software I', 'Base de Datos I'),
    ('Ingeniería de Software I', 'Metodologías de Desarrollo de Software'),
    -- 4° año
    ('Organización y Gestión Empresarial', 'Metodologías de Desarrollo de Software'),
    ('Organización y Gestión Empresarial', 'Probabilidad y Estadística'),
    ('Teoría de la Información', 'Probabilidad y Estadística'),
    ('Teoría de la Información', 'Álgebra Lineal'),
    ('Redes de Computadoras II', 'Redes de Computadoras I'),
    ('Física II', 'Física I'),
    ('Física II', 'Cálculo II'),
    ('Calidad de Software', 'Ingeniería de Software I'),
    ('Arquitectura de Computadoras II', 'Arquitectura de Computadoras I'),
    ('Fundamentos de la Ciencia de Datos', 'Probabilidad y Estadística'),
    ('Fundamentos de la Ciencia de Datos', 'Base de Datos I'),
    ('Compiladores e Intérpretes', 'Lenguajes y Paradigmas'),
    -- 5° año
    ('Ingeniería de Software II', 'Calidad de Software'),
    ('Ciberseguridad', 'Redes de Computadoras I'),
    ('Ciberseguridad', 'Probabilidad y Estadística'),
    ('Ciberseguridad', 'Análisis y Diseño de Algoritmos II'),
    ('Inteligencia Artificial', 'Fundamentos de la Ciencia de Datos'),
    ('Inteligencia Artificial', 'Probabilidad y Estadística'),
    ('Inteligencia Artificial', 'Análisis y Diseño de Algoritmos II'),
    ('Formulación y Evaluación de Proyectos TICs', 'Organización y Gestión Empresarial'),
    ('Ética y Legislación de la Práctica Profesional', 'Organización y Gestión Empresarial'),
    ('Práctica Profesional Supervisada y Proyecto Integrador', 'Ingeniería de Software II'),
    ('Práctica Profesional Supervisada y Proyecto Integrador', 'Ciberseguridad'),
    ('Práctica Profesional Supervisada y Proyecto Integrador', 'Inteligencia Artificial')
) AS c(materia, requerida)
JOIN materia m ON m.nombre = c.materia
JOIN materia r ON r.nombre = c.requerida
ON CONFLICT DO NOTHING;

-- Alumno de prueba (la página muestra siempre su plan)
INSERT INTO usuario (nombre_apellido, email, telefono)
VALUES ('Alumno de Prueba', 'alumno@prueba.com', '000000000')
ON CONFLICT (email) DO NOTHING;

-- Cursadas iniciales del alumno de prueba: alguien que está en 2° año.
-- Respetan las correlatividades:
--   finalizada -> sus correlativas finalizadas
--   regular / cursando -> sus correlativas regulares o finalizadas
INSERT INTO usuario_materia (usuario_id_usuario, materia_id_materia, estado, nota)
SELECT u.id_usuario, m.id_materia, c.estado, c.nota
FROM (VALUES
    -- 1° año, 1er cuatrimestre: todo finalizado
    ('Introducción a la Programación I', 'finalizada', 8::real),
    ('Introducción al Álgebra', 'finalizada', 7),
    ('Introducción a los Sistemas Informáticos', 'finalizada', 9),
    ('Inglés', 'finalizada', 6),
    -- 1° año, 2do cuatrimestre: dos finalizadas, dos regulares (final pendiente)
    ('Diseño Lógico', 'finalizada', 7),
    ('Introducción a la Programación II', 'finalizada', 8),
    ('Lenguajes Formales y Autómatas', 'regular', 6),
    ('Cálculo I', 'regular', 5),
    -- 2° año, 1er cuatrimestre: cursando ahora
    ('Arquitectura de Computadoras I', 'cursando', NULL),
    ('Análisis y Diseño de Algoritmos I', 'cursando', NULL),
    ('Programación Orientada a Objetos', 'cursando', NULL),
    ('Álgebra Lineal', 'cursando', NULL)
) AS c(materia, estado, nota)
JOIN materia m ON m.nombre = c.materia
JOIN usuario u ON u.email = 'alumno@prueba.com'
ON CONFLICT DO NOTHING;
