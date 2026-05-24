package services

import (
	"context"
	"strconv"
	"strings"
	"time"

	"skihub/internal/infonieve"
	"skihub/internal/models"
	"skihub/internal/repository"
)

// EstacionService centraliza la logica relacionada con estaciones.
type EstacionService struct {
	Repo    *repository.EstacionRepo
	FavRepo *repository.FavoritoRepo
	// NieveSvc enriquece las estaciones con datos en tiempo real de infonieve.es.
	// Si es nil, los campos dinamicos quedan a cero/vacio (nunca ficticios).
	NieveSvc *NieveService
}

// NuevoEstacionService construye el servicio. Si favRepo es nil, no se marcan favoritas.
func NuevoEstacionService(repo *repository.EstacionRepo, favRepo *repository.FavoritoRepo) *EstacionService {
	return &EstacionService{Repo: repo, FavRepo: favRepo}
}

// Listar devuelve todas las estaciones ordenadas por distancia. Si usuarioID > 0
// marca las que son favoritas. Los campos dinamicos se enriquecen con datos reales.
func (s *EstacionService) Listar(ctx context.Context, usuarioID int64) ([]models.Estacion, error) {
	lista, err := s.Repo.ListarPorDistancia(ctx)
	if err != nil {
		return nil, err
	}
	if usuarioID > 0 && s.FavRepo != nil {
		ids, err := s.FavRepo.IDsDeUsuario(ctx, usuarioID)
		if err != nil {
			return nil, err
		}
		for i := range lista {
			if ids[lista[i].ID] {
				lista[i].EsFavorita = true
			}
		}
	}
	for i := range lista {
		s.enriquecerConNieve(&lista[i])
	}
	return lista, nil
}

// Obtener devuelve la ficha de una estacion, marcandola como favorita si
// procede. Los campos dinamicos se enriquecen con datos reales de infonieve.es.
func (s *EstacionService) Obtener(ctx context.Context, id, usuarioID int64) (*models.Estacion, error) {
	e, err := s.Repo.ObtenerPorID(ctx, id)
	if err != nil {
		return nil, err
	}
	if usuarioID > 0 && s.FavRepo != nil {
		fav, err := s.FavRepo.Existe(ctx, usuarioID, id)
		if err != nil {
			return nil, err
		}
		e.EsFavorita = fav
	}
	s.enriquecerConNieve(e)
	return e, nil
}

// ResumenHome calcula la estacion mas cercana, la mas lejana y la
// distancia promedio para el panel de estadisticas de la home.
func (s *EstacionService) ResumenHome(ctx context.Context) (cercana, lejana *models.Estacion, promedio float64, total int, err error) {
	lista, err := s.Repo.ListarPorDistancia(ctx)
	if err != nil || len(lista) == 0 {
		return nil, nil, 0, 0, err
	}
	var suma float64
	for i := range lista {
		suma += lista[i].Distancia
	}
	c := lista[0]
	l := lista[len(lista)-1]
	s.enriquecerConNieve(&c)
	s.enriquecerConNieve(&l)
	return &c, &l, suma / float64(len(lista)), len(lista), nil
}

// enriquecerConNieve reemplaza a enriquecerParte(). Obtiene datos en tiempo real
// de infonieve.es para la estacion e a traves de NieveSvc.
//
// Si hay datos reales: sobrescribe los campos dinamicos (pistas, remontes,
// km, nieve, temperatura) con los valores de infonieve.es y marca
// TieneDatosReales=true.
//
// Si NO hay datos reales (servicio nil, estacion sin mapping, o error de red):
// llama a limpiarCamposDinamicos — nunca se generan valores ficticios.
func (s *EstacionService) enriquecerConNieve(e *models.Estacion) {
	if s.NieveSvc == nil {
		limpiarCamposDinamicos(e)
		return
	}
	dto := s.NieveSvc.PorNombre(e.Nombre)
	if dto == nil {
		limpiarCamposDinamicos(e)
		return
	}

	// Pistas abiertas / totales
	if dto.Pistas.Abiertos != nil {
		e.PistasAbiertas = int(*dto.Pistas.Abiertos)
	} else if dto.Estado == infonieve.EstadoCerrada {
		e.PistasAbiertas = 0
	}
	if dto.Pistas.Total != nil {
		if t := int(*dto.Pistas.Total); t > 0 {
			e.PistasTotales = t
		}
	}

	// Remontes
	if dto.Remontes.Abiertos != nil {
		e.RemontesOp = int(*dto.Remontes.Abiertos)
	}
	if dto.Remontes.Total != nil {
		if t := int(*dto.Remontes.Total); t > 0 {
			e.RemontesTot = t
		}
	}

	// Km esquiables
	if dto.Kilometros.Abiertos != nil {
		e.KmEsquiables = *dto.Kilometros.Abiertos
	}

	// Nieve: infonieve da un unico valor; lo asignamos a Min y Max.
	// Si es nil (sin datos de nieve) dejamos ambos a 0.
	if dto.NieveCm != nil {
		cm := int(*dto.NieveCm)
		e.NieveBase = cm
		e.NieveMin = cm
		e.NieveMax = cm
	} else {
		e.NieveMin = 0
		e.NieveMax = 0
	}

	// Temperatura: infonieve la devuelve como string "2C", "-5 C", etc.
	if dto.Temperatura != "" {
		e.Temperatura = parsearTemperatura(dto.Temperatura)
	}

	// Viento: no esta disponible en el listado de infonieve -> vacio, no ficticio.
	e.Viento = ""

	// Timestamp del parte (cache ~10 min; usamos now() como aproximacion).
	e.ParteActualizado = time.Now()

	e.TieneDatosReales = true
}

// limpiarCamposDinamicos pone a cero/vacio todos los campos que antes
// enriquecerParte() generaba de forma ficticia desde un hash del ID.
func limpiarCamposDinamicos(e *models.Estacion) {
	e.NieveMin = 0
	e.NieveMax = 0
	e.Viento = ""
	e.ParteActualizado = time.Time{} // zero -> ParteHora() devuelve ""
	e.TieneDatosReales = false
}

// parsearTemperatura extrae el entero con signo de cadenas como
// "2C", "-5 C", "+3C". Devuelve 0 si no puede parsear.
func parsearTemperatura(s string) int {
	s = strings.TrimSpace(s)
	var sign, digits string
	i := 0
	if i < len(s) && (s[i] == '-' || s[i] == '+') {
		if s[i] == '-' {
			sign = "-"
		}
		i++
	}
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		digits += string(s[i])
		i++
	}
	if digits == "" {
		return 0
	}
	n, _ := strconv.Atoi(sign + digits)
	return n
}
