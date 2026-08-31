CREATE TABLE usuario (
    id_usuario SERIAL NOT NULL,
    nombre_apellido VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL,
    telefono VARCHAR(20) NOT NULL,
    CONSTRAINT pk_usuario PRIMARY KEY (id_usuario),
    CONSTRAINT uq_usuario_email UNIQUE (email)
);

CREATE TABLE materia (
    id_materia SERIAL NOT NULL,
    nombre VARCHAR(100) NOT NULL,
    anio INT NOT NULL,
    cuatrimestre INT NOT NULL,
    CONSTRAINT pk_materia PRIMARY KEY (id_materia)
);

CREATE TABLE usuario_materia (
    usuario_id_usuario INT NOT NULL,
    materia_id_materia INT NOT NULL,
    estado VARCHAR(30) NOT NULL,
    nota REAL NULL,
    CONSTRAINT pk_usuario_materia PRIMARY KEY (usuario_id_usuario, materia_id_materia)
);

CREATE TABLE correlatividad_materia (
    materia_id_materia INT NOT NULL,
    id_materia_requerida INT NOT NULL,
    CONSTRAINT pk_correlatividad_materia PRIMARY KEY (materia_id_materia, id_materia_requerida)
);

ALTER TABLE usuario_materia
    ADD CONSTRAINT fk_usuario_materia_usuario
    FOREIGN KEY (usuario_id_usuario)
    REFERENCES usuario (id_usuario)
    ON DELETE CASCADE;

ALTER TABLE usuario_materia
    ADD CONSTRAINT fk_usuario_materia_materia
    FOREIGN KEY (materia_id_materia)
    REFERENCES materia (id_materia)
    ON DELETE CASCADE;

ALTER TABLE correlatividad_materia
    ADD CONSTRAINT fk_correlatividad_materia_principal
    FOREIGN KEY (materia_id_materia)
    REFERENCES materia (id_materia)
    ON DELETE CASCADE;

ALTER TABLE correlatividad_materia
    ADD CONSTRAINT fk_correlatividad_materia_requerida
    FOREIGN KEY (id_materia_requerida)
    REFERENCES materia (id_materia)
    ON DELETE CASCADE;
