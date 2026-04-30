// Package infonieve es un cliente Go para extraer datos en tiempo real
// del estado de las estaciones de esquí de infonieve.es.
//
// Esta es una reimplementación en Go puro del scraper original que se
// distribuye como `infonieve-api` (Node.js). Mantenemos los nombres y
// la estructura de los datos para que el formato JSON expuesto por
// `/api/nieve/...` sea estable.
//
// El cliente HTTP reutiliza conexiones (keep-alive), aplica un timeout
// razonable y reintenta con backoff exponencial ante errores transitorios.
// No reintenta ante 404 ni 429.
package infonieve

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// BaseURL es el dominio raíz desde donde se obtienen los datos. Se puede
// sobrescribir en tests apuntándolo a un servidor local.
const BaseURL = "https://www.infonieve.es"

// Cabeceras "humanas" para que infonieve.es no rechace al cliente.
var defaultHeaders = map[string]string{
	"User-Agent":      "Mozilla/5.0 (compatible; Snowbreak/1.0; +https://snowbreak.local)",
	"Accept-Language": "es-ES,es;q=0.9,en;q=0.8",
	"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
}

// Errores tipados para que las capas superiores puedan reaccionar de forma
// específica (devolver 404 al cliente, mostrar "rate-limited", etc.).
var (
	ErrNotFound    = errors.New("infonieve: recurso no encontrado")
	ErrRateLimited = errors.New("infonieve: peticiones limitadas (429)")
	ErrTimeout     = errors.New("infonieve: timeout")
	ErrNetwork     = errors.New("infonieve: error de red")
)

// Client encapsula al http.Client. Es seguro para uso concurrente.
type Client struct {
	http    *http.Client
	base    string
	retries int
}

// NewClient crea un cliente listo para producción con keep-alive.
func NewClient() *Client {
	return &Client{
		http: &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        20,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		base:    BaseURL,
		retries: 2,
	}
}

// WithBaseURL cambia la URL base (útil para tests con servidor local).
func (c *Client) WithBaseURL(base string) *Client {
	c.base = base
	return c
}

// fetch descarga la URL devolviendo el HTML como []byte. Aplica retries.
func (c *Client) fetch(rawURL string) ([]byte, error) {
	if _, err := url.Parse(rawURL); err != nil {
		return nil, fmt.Errorf("URL inválida: %w", err)
	}

	var lastErr error
	for intento := 0; intento <= c.retries; intento++ {
		body, err := c.doRequest(rawURL)
		if err == nil {
			return body, nil
		}
		lastErr = err

		// No reintentamos en 404 ni 429.
		if errors.Is(err, ErrNotFound) || errors.Is(err, ErrRateLimited) {
			return nil, err
		}
		// Backoff exponencial: 500ms, 1000ms.
		if intento < c.retries {
			time.Sleep(time.Duration(500*(1<<intento)) * time.Millisecond)
		}
	}
	return nil, lastErr
}

// doRequest realiza una sola petición y mapea los errores HTTP a errores
// tipados de paquete.
func (c *Client) doRequest(rawURL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetwork, err)
	}
	for k, v := range defaultHeaders {
		req.Header.Set(k, v)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		// http.Client devuelve un error que envuelve un context.DeadlineExceeded
		// cuando salta el timeout; lo detectamos por el texto del mensaje
		// para no acoplarnos a la implementación interna.
		if isTimeoutError(err) {
			return nil, fmt.Errorf("%w: %v", ErrTimeout, err)
		}
		return nil, fmt.Errorf("%w: %v", ErrNetwork, err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return nil, ErrNotFound
	case resp.StatusCode == http.StatusTooManyRequests:
		return nil, ErrRateLimited
	case resp.StatusCode >= 500:
		return nil, fmt.Errorf("%w: HTTP %d", ErrNetwork, resp.StatusCode)
	case resp.StatusCode >= 400:
		return nil, fmt.Errorf("%w: HTTP %d", ErrNetwork, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: lectura body: %v", ErrNetwork, err)
	}
	return body, nil
}

// isTimeoutError detecta errores derivados de timeout sin importar
// la versión exacta del paquete net/http.
func isTimeoutError(err error) bool {
	type timeoutter interface{ Timeout() bool }
	var t timeoutter
	if errors.As(err, &t) {
		return t.Timeout()
	}
	return false
}
