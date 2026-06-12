package services

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"skihub/internal/models"
	"skihub/internal/repository"
)

// Errores de validación del módulo de hoteles. Los handlers los comparan
// con errors.Is para mostrar mensajes amigables.
var (
	ErrHotelNombreVacio   = errors.New("el nombre del hotel es obligatorio")
	ErrHotelSinEstacion   = errors.New("el hotel debe estar asociado a una estación")
	ErrHotelURLInvalida   = errors.New("URL externa inválida: debe ser https:// y un dominio válido")
	ErrHotelRatingInvalid = errors.New("la valoración debe estar entre 0 y 5")
	ErrHotelPrecioInvalid = errors.New("el precio y la distancia no pueden ser negativos")
)

// HotelService centraliza la lógica del módulo de hoteles: validación de
// campos y URLs externas, generación de slugs y acceso a la galería.
type HotelService struct {
	Repo *repository.HotelRepo
}

// NuevoHotelService construye el servicio.
func NuevoHotelService(repo *repository.HotelRepo) *HotelService {
	return &HotelService{Repo: repo}
}

// Listar devuelve los hoteles según el filtro dado.
func (s *HotelService) Listar(ctx context.Context, f repository.FiltroHoteles) ([]models.Hotel, error) {
	return s.Repo.Listar(ctx, f)
}

// ObtenerPorSlug devuelve la ficha de un hotel con su galería cargada.
func (s *HotelService) ObtenerPorSlug(ctx context.Context, slug string) (*models.Hotel, error) {
	h, err := s.Repo.ObtenerPorSlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	h.Imagenes, err = s.Repo.ImagenesDeHotel(ctx, h.ID)
	if err != nil {
		return nil, err
	}
	return h, nil
}

// ObtenerPorID devuelve la ficha de un hotel con su galería (panel admin).
func (s *HotelService) ObtenerPorID(ctx context.Context, id int64) (*models.Hotel, error) {
	h, err := s.Repo.ObtenerPorID(ctx, id)
	if err != nil {
		return nil, err
	}
	h.Imagenes, err = s.Repo.ImagenesDeHotel(ctx, h.ID)
	if err != nil {
		return nil, err
	}
	return h, nil
}

// Guardar valida y persiste un hotel (crea si ID==0, actualiza si no).
// Devuelve el id final. `urlsGaleria` reemplaza la galería completa
// (una URL por elemento; la primera pasa a ser la imagen primaria).
func (s *HotelService) Guardar(ctx context.Context, h *models.Hotel, urlsGaleria []string) (int64, error) {
	if err := s.validar(h); err != nil {
		return 0, err
	}
	if h.Slug == "" {
		slug, err := s.slugLibre(ctx, h.Nombre, h.ID)
		if err != nil {
			return 0, err
		}
		h.Slug = slug
	}

	imgs, err := validarURLsGaleria(urlsGaleria, h.Nombre)
	if err != nil {
		return 0, err
	}

	id := h.ID
	if id == 0 {
		id, err = s.Repo.Crear(ctx, h)
	} else {
		err = s.Repo.Actualizar(ctx, h)
	}
	if err != nil {
		return 0, err
	}
	if err := s.Repo.ReemplazarImagenes(ctx, id, imgs); err != nil {
		return 0, err
	}
	return id, nil
}

// Borrar elimina un hotel y su galería.
func (s *HotelService) Borrar(ctx context.Context, id int64) error {
	return s.Repo.Borrar(ctx, id)
}

// validar comprueba campos obligatorios, rangos y URLs externas.
func (s *HotelService) validar(h *models.Hotel) error {
	h.Nombre = strings.TrimSpace(h.Nombre)
	if h.Nombre == "" {
		return ErrHotelNombreVacio
	}
	if h.EstacionID <= 0 {
		return ErrHotelSinEstacion
	}
	if h.Rating < 0 || h.Rating > 5 {
		return ErrHotelRatingInvalid
	}
	if h.PrecioDesde < 0 || h.DistanciaEstacion < 0 || h.NumResenas < 0 {
		return ErrHotelPrecioInvalid
	}
	if h.Divisa == "" {
		h.Divisa = "EUR"
	}
	// URLs externas: o vacías o https válidas (anti-XSS: nada de javascript:).
	for _, campo := range []*string{&h.BookingURL, &h.OficialURL, &h.ImagenURL} {
		limpio, ok := SanearURLExterna(*campo)
		if !ok {
			return ErrHotelURLInvalida
		}
		*campo = limpio
	}
	return nil
}

// SanearURLExterna valida una URL externa. Devuelve la URL normalizada y
// true si es aceptable. Reglas: vacía es válida (campo opcional); si no,
// debe parsear, ser https y tener host. Cualquier otro esquema
// (javascript:, data:, http:, ftp:...) se rechaza.
func SanearURLExterna(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", true
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", false
	}
	// Rutas locales a /static (imágenes alojadas por nosotros) son válidas.
	if u.Scheme == "" && u.Host == "" && strings.HasPrefix(u.Path, "/static/") {
		return u.String(), true
	}
	if !strings.EqualFold(u.Scheme, "https") || u.Host == "" {
		return "", false
	}
	return u.String(), true
}

// validarURLsGaleria convierte las líneas del formulario admin en imágenes
// validadas. Las vacías se ignoran; una URL inválida aborta el guardado.
func validarURLsGaleria(urls []string, nombreHotel string) ([]models.HotelImagen, error) {
	imgs := make([]models.HotelImagen, 0, len(urls))
	for _, raw := range urls {
		limpio, ok := SanearURLExterna(raw)
		if !ok {
			return nil, fmt.Errorf("%w: %q", ErrHotelURLInvalida, strings.TrimSpace(raw))
		}
		if limpio == "" {
			continue
		}
		imgs = append(imgs, models.HotelImagen{
			URL:     limpio,
			AltText: "Foto de " + nombreHotel,
		})
	}
	return imgs, nil
}

// GenerarSlug convierte un nombre en slug URL-safe: minúsculas, sin tildes,
// solo [a-z0-9-]. "Hôtel Pic d'Or 4★" → "hotel-pic-d-or-4".
func GenerarSlug(nombre string) string {
	s := strings.ToLower(strings.TrimSpace(nombre))
	r := strings.NewReplacer(
		"á", "a", "à", "a", "ä", "a", "â", "a",
		"é", "e", "è", "e", "ë", "e", "ê", "e",
		"í", "i", "ì", "i", "ï", "i", "î", "i",
		"ó", "o", "ò", "o", "ö", "o", "ô", "o",
		"ú", "u", "ù", "u", "ü", "u", "û", "u",
		"ñ", "n", "ç", "c",
	)
	s = r.Replace(s)
	var b strings.Builder
	prevGuion := false
	for _, c := range s {
		switch {
		case (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9'):
			b.WriteRune(c)
			prevGuion = false
		default:
			if !prevGuion && b.Len() > 0 {
				b.WriteByte('-')
				prevGuion = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// slugLibre genera un slug único: si ya existe, prueba con sufijos -2, -3…
func (s *HotelService) slugLibre(ctx context.Context, nombre string, excluirID int64) (string, error) {
	base := GenerarSlug(nombre)
	if base == "" {
		base = "hotel"
	}
	slug := base
	for i := 2; i <= 50; i++ {
		existe, err := s.Repo.ExisteSlug(ctx, slug, excluirID)
		if err != nil {
			return "", err
		}
		if !existe {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
	return "", fmt.Errorf("no se pudo generar slug único para %q", nombre)
}
