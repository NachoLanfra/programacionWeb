// URL base de la API. Es relativa porque el HTML lo sirve el mismo servidor Go.
const API_URL = '/materias';

// Referencias a los elementos del DOM que vamos a usar.
const lista = document.getElementById('lista-materias');
const mensaje = document.getElementById('mensaje');

// Pide todas las materias a la API y las dibuja en la lista.
async function cargarMaterias() {
    try {
        const res = await fetch(API_URL);
        if (!res.ok) {
            throw new Error(`Error ${res.status} al pedir las materias`);
        }
        const materias = await res.json();
        renderMaterias(materias);
    } catch (err) {
        console.error(err);
        mensaje.textContent = 'No se pudieron cargar las materias.';
    }
}

// Genera un <li> por cada materia y los agrega al contenedor.
function renderMaterias(materias) {
    // Vaciamos la lista para no duplicar elementos al refrescar.
    lista.replaceChildren();

    if (materias.length === 0) {
        const vacio = document.createElement('li');
        vacio.textContent = 'No hay materias cargadas.';
        lista.append(vacio);
        return;
    }

    for (const materia of materias) {
        const li = document.createElement('li');

        const info = document.createElement('span');
        info.textContent = `${materia.nombre} - ${materia.anio}° año, ${materia.cuatrimestre}° cuatrimestre`;

        const botonEliminar = document.createElement('button');
        botonEliminar.type = 'button';
        botonEliminar.className = 'btn-eliminar';
        botonEliminar.dataset.id = materia.id_materia;
        botonEliminar.textContent = 'Eliminar';

        li.append(info, botonEliminar);
        lista.append(li);
    }
}

// Arma un texto legible a partir del cuerpo de error que devuelve la API.
// Ejemplo de cuerpo: {"error": "error de validación", "campos": {"anio": "debe ser entre 1 y 5"}}
function textoDeError(cuerpo) {
    if (cuerpo.campos) {
        return Object.entries(cuerpo.campos)
            .map(([campo, detalle]) => `${campo}: ${detalle}`)
            .join(', ');
    }
    return cuerpo.error || 'Ocurrió un error inesperado.';
}

const form = document.getElementById('form-materia');

// Al enviar el formulario: cancelamos la recarga, mandamos el POST y refrescamos la lista.
form.addEventListener('submit', async (event) => {
    event.preventDefault();

    // Los inputs siempre devuelven texto, así que convertimos los números.
    const datos = {
        nombre: form.nombre.value.trim(),
        anio: Number(form.anio.value),
        cuatrimestre: Number(form.cuatrimestre.value),
    };

    try {
        const res = await fetch(API_URL, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(datos),
        });

        if (!res.ok) {
            const cuerpo = await res.json().catch(() => ({}));
            mensaje.textContent = textoDeError(cuerpo);
            return;
        }

        form.reset();
        mensaje.textContent = 'Materia agregada.';
        await cargarMaterias();
    } catch (err) {
        console.error(err);
        mensaje.textContent = 'No se pudo conectar con el servidor.';
    }
});

// Delegación de eventos: un solo listener en la lista atiende los clics
// de todos los botones Eliminar, incluso los que se crean después.
lista.addEventListener('click', async (event) => {
    const boton = event.target.closest('.btn-eliminar');
    if (!boton) {
        return;
    }

    const id = boton.dataset.id;

    try {
        const res = await fetch(`${API_URL}/${id}`, { method: 'DELETE' });

        // El backend responde 204 sin cuerpo, así que no leemos JSON si salió bien.
        if (!res.ok) {
            const cuerpo = await res.json().catch(() => ({}));
            mensaje.textContent = textoDeError(cuerpo);
            return;
        }

        mensaje.textContent = 'Materia eliminada.';
        await cargarMaterias();
    } catch (err) {
        console.error(err);
        mensaje.textContent = 'No se pudo conectar con el servidor.';
    }
});

// Carga inicial. El script tiene defer, así que el DOM ya existe en este punto.
cargarMaterias();
