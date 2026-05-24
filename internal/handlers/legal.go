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
		Titulo:      "Aviso legal - SnowBreak",
		Descripcion: "Información legal de SnowBreak.",
		Heading:     "Aviso legal",
		Intro:       "Información legal sobre Snowbreak y el alcance de sus servicios.",
		Sections: []legalSection{
			{
				Title: "Titularidad del sitio",
				Paragraphs: []string{
					"Snowbreak es una plataforma de información turística especializada en viajes de esquí y deportes de nieve.",
					"El contenido se ofrece con finalidad informativa. Los precios y disponibilidades son orientativos y están sujetos a confirmación.",
				},
			},
			{
				Title: "Uso de la información",
				Paragraphs: []string{
					"Los datos mostrados sobre estaciones, condiciones de nieve y noticias tienen carácter informativo y orientativo.",
					"No debe tomarse ninguna decisión de viaje, seguridad o compra basándose exclusivamente en esta web.",
				},
			},
		},
		Usuario: a.UsuarioActual(r),
	})
}

func (a *App) PoliticaPrivacidad(w http.ResponseWriter, r *http.Request) {
	render(w, r, a.Plantillas, "legal", datosLegal{
		Titulo:      "Política de privacidad - SnowBreak",
		Descripcion: "Política de privacidad de SnowBreak.",
		Heading:     "Política de privacidad",
		Intro:       "Información sobre el tratamiento de datos personales en Snowbreak.",
		Sections: []legalSection{
			{
				Title: "Datos tratados",
				Paragraphs: []string{
					"El sitio puede almacenar nombre, correo electrónico y credenciales cifradas para permitir el registro y acceso de usuarios.",
					"Estos datos se utilizan para gestionar tu cuenta, tus estaciones favoritas y el acceso a los servicios de la plataforma.",
				},
			},
			{
				Title: "Finalidad y conservación",
				Paragraphs: []string{
					"Los datos se conservan mientras la cuenta esté activa. Puedes solicitar su eliminación en cualquier momento contactando con nosotros.",
					"No se comparten con terceros ni se destinan a fines publicitarios.",
				},
			},
		},
		Usuario: a.UsuarioActual(r),
	})
}

func (a *App) PoliticaCookies(w http.ResponseWriter, r *http.Request) {
	render(w, r, a.Plantillas, "legal", datosLegal{
		Titulo:      "Política de cookies - SnowBreak",
		Descripcion: "Política de cookies de SnowBreak.",
		Heading:     "Política de cookies",
		Intro:       "Uso de cookies técnicas dentro de la aplicación.",
		Sections: []legalSection{
			{
				Title: "Cookies necesarias",
				Paragraphs: []string{
					"SnowBreak utiliza una cookie de sesión para mantener al usuario autenticado tras iniciar sesión.",
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
