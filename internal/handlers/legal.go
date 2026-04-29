package handlers

import "net/http"

type legalSection struct {
	Title      string
	Paragraphs []string
}

type datosLegal struct {
	Titulo      string
	Descripcion string
	Activa      string
	Heading     string
	Intro       string
	Sections    []legalSection
	Usuario     any
}

func (a *App) AvisoLegal(w http.ResponseWriter, r *http.Request) {
	render(w, r, a.Plantillas, "legal", datosLegal{
		Titulo:      "Aviso legal - SkiHub",
		Descripcion: "Información legal de SkiHub.",
		Heading:     "Aviso legal",
		Intro:       "Información básica sobre este sitio web académico y su alcance.",
		Sections: []legalSection{
			{
				Title: "Titularidad del sitio",
				Paragraphs: []string{
					"SkiHub es un proyecto académico desarrollado con fines docentes dentro de la asignatura de Sistemas Web.",
					"El contenido del sitio se ofrece como demostración técnica y no constituye un servicio comercial real.",
				},
			},
			{
				Title: "Uso de la información",
				Paragraphs: []string{
					"Los datos mostrados sobre estaciones, condiciones o noticias tienen carácter informativo y de ejemplo.",
					"No debe tomarse ninguna decisión de viaje, seguridad o compra basándose exclusivamente en esta web.",
				},
			},
		},
		Usuario: a.UsuarioActual(r),
	})
}

func (a *App) PoliticaPrivacidad(w http.ResponseWriter, r *http.Request) {
	render(w, r, a.Plantillas, "legal", datosLegal{
		Titulo:      "Política de privacidad - SkiHub",
		Descripcion: "Política de privacidad de SkiHub.",
		Heading:     "Política de privacidad",
		Intro:       "Resumen del tratamiento de datos dentro de este proyecto académico.",
		Sections: []legalSection{
			{
				Title: "Datos tratados",
				Paragraphs: []string{
					"El sitio puede almacenar nombre, correo electrónico y credenciales cifradas para permitir el registro y acceso de usuarios.",
					"Estos datos se utilizan únicamente para demostrar funcionalidades de autenticación, favoritos y administración.",
				},
			},
			{
				Title: "Finalidad y conservación",
				Paragraphs: []string{
					"Los datos se conservan mientras exista la base de datos usada en el entorno de prácticas.",
					"No se comparten con terceros ni se destinan a fines publicitarios.",
				},
			},
		},
		Usuario: a.UsuarioActual(r),
	})
}

func (a *App) PoliticaCookies(w http.ResponseWriter, r *http.Request) {
	render(w, r, a.Plantillas, "legal", datosLegal{
		Titulo:      "Política de cookies - SkiHub",
		Descripcion: "Política de cookies de SkiHub.",
		Heading:     "Política de cookies",
		Intro:       "Uso de cookies técnicas dentro de la aplicación.",
		Sections: []legalSection{
			{
				Title: "Cookies necesarias",
				Paragraphs: []string{
					"SkiHub utiliza una cookie de sesión para mantener al usuario autenticado tras iniciar sesión.",
					"Esta cookie es técnica y necesaria para funcionalidades como favoritos, cambio de contraseña y panel de administración.",
				},
			},
			{
				Title: "Cookies de terceros",
				Paragraphs: []string{
					"El sitio no incorpora herramientas analíticas ni cookies publicitarias propias.",
					"Al cargar el mapa pueden solicitarse recursos de terceros para mostrar la cartografía base.",
				},
			},
		},
		Usuario: a.UsuarioActual(r),
	})
}
