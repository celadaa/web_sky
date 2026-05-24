package handlers

import (
	"log"
	"net/http"

	"skihub/internal/models"
)

// datosHome es el struct que se pasa a la plantilla index.tmpl.
type datosHome struct {
	Titulo                string
	Descripcion           string
	Activa                string
	Estaciones            []models.Estacion // todas (autocomplete + métricas)
	EstacionesDestacadas  []models.Estacion // solo las 3 de la home
	MasCercana            *models.Estacion
	MasLejana             *models.Estacion
	DistanciaPromedio     float64
	Usuario               *models.Usuario
}

// Home responde a GET / con el listado de estaciones más cercanas y el resumen.
func (a *App) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		a.NotFound(w, r)
		return
	}
	ctx := r.Context()
	cercana, lejana, promedio, _, err := a.EstacionSvc.ResumenHome(ctx)
	if err != nil {
		log.Printf("ERROR resumen home: %v", err)
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	u := a.UsuarioActual(r)
	var uid int64
	if u != nil {
		uid = u.ID
	}
	estaciones, err := a.EstacionSvc.Listar(ctx, uid)
	if err != nil {
		log.Printf("ERROR listar estaciones: %v", err)
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}

	// Estaciones destacadas en la home: solo Baqueira, Sierra Nevada y Candanchú.
	// El resto se mantiene en Estaciones para el autocomplete del buscador.
	nombresDestacadas := map[string]int{
		"Baqueira Beret": 0,
		"Sierra Nevada":  1,
		"Candanchú":      2,
	}
	destacadas := make([]models.Estacion, 3)
	found := 0
	for _, e := range estaciones {
		if pos, ok := nombresDestacadas[e.Nombre]; ok {
			destacadas[pos] = e
			found++
			if found == 3 {
				break
			}
		}
	}
	// Eliminar posiciones vacías si alguna estación no existe en la BD
	var destacadasFinal []models.Estacion
	for _, e := range destacadas {
		if e.ID != 0 {
			destacadasFinal = append(destacadasFinal, e)
		}
	}
	// Fallback: si no se encontró ninguna, usar las 3 primeras
	if len(destacadasFinal) == 0 && len(estaciones) > 0 {
		n := 3
		if len(estaciones) < n {
			n = len(estaciones)
		}
		destacadasFinal = estaciones[:n]
	}

	render(w, r, a.Plantillas, "index", datosHome{
		Titulo:               "SnowBreak | Estaciones de esquí y forfaits",
		Descripcion:          "Compara estaciones de esquí, consulta el estado de pistas y encuentra forfaits para tu próxima escapada a la nieve con SnowBreak.",
		Activa:               "inicio",
		Estaciones:           estaciones,
		EstacionesDestacadas: destacadasFinal,
		MasCercana:           cercana,
		MasLejana:            lejana,
		DistanciaPromedio:    promedio,
		Usuario:              u,
	})
}
