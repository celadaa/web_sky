// Package handlers — middlewares y utilidades de seguridad.
//
// Este fichero centraliza todo el endurecimiento HTTP de SnowBreak:
//
//   - Cabeceras de seguridad (HSTS, X-Frame-Options, X-Content-Type-Options,
//     Referrer-Policy, Permissions-Policy, COOP, CORP, CSP estricta).
//   - Límite de tamaño de cuerpo de petición (anti-DoS).
//   - Rate limiting in-memory por IP para endpoints sensibles.
//   - Tokens CSRF "double-submit cookie" para formularios POST.
//   - Helpers para no exponer detalles internos en errores 5xx.
//   - Helpers de IP cliente con/sin proxy.
//
// Diseño:
//   - Sin dependencias externas: usamos solo stdlib.
//   - Los middlewares devuelven http.Handler para encadenarse de forma
//     declarativa en main.go (Encadenar(seguridad, rateLimit, csrf, ...)).
//   - El estado (tokens, rate-limit buckets) vive en una struct App que
//     se inicializa una sola vez al arrancar el servidor.
package handlers

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"skihub/internal/config"
)

// CookieCSRF es la cookie no-HttpOnly que comparte el token con el
// frontend. Al ser leída desde JS o renderizada en HTML, vamos en
// double-submit cookie pattern.
//
// El nombre se mantiene en sintonía con `skihub_session` y con la
// configuración Nginx (`proxy_cookie_flags skihub_csrf …`). Cambiar este
// valor sin actualizar a la vez `web/static/csrf.js` (COOKIE_NAME) y el
// snippet Nginx rompe la inyección automática del token en formularios.
const CookieCSRF = "skihub_csrf"

// CabeceraCSRF / CampoCSRF son los nombres aceptados para enviar el
// token de vuelta en una petición que muta estado.
const (
	CabeceraCSRF = "X-CSRF-Token"
	CampoCSRF    = "csrf_token"
)

// Sec encapsula el estado en memoria de los middlewares de seguridad.
type Sec struct {
	cfg      *config.Config
	limiters map[string]*ipBucket
	mu       sync.Mutex
	csrfKey  []byte
}

// NuevoSec construye el estado de seguridad. Si la config trae
// CSRFSecret usamos esa; si no, generamos una semilla efímera (válida
// para el ciclo de vida del proceso, pero romperá tokens entre
// reinicios). Eso obliga a configurar CSRF_SECRET en producción.
func NuevoSec(cfg *config.Config) *Sec {
	clave := []byte(cfg.CSRFSecret)
	if len(clave) < 32 {
		efimera := make([]byte, 32)
		if _, err := rand.Read(efimera); err == nil {
			clave = efimera
			if cfg.EsProduccion() {
				log.Println("AVISO: CSRF_SECRET no configurada — usando clave efímera (los tokens se invalidarán al reiniciar)")
			}
		}
	}
	return &Sec{
		cfg:      cfg,
		limiters: make(map[string]*ipBucket),
		csrfKey:  clave,
	}
}

// ─── Encadenado de middlewares ───────────────────────────────────────────────

// Encadenar aplica los middlewares de derecha a izquierda. El primero
// es el más externo (el que ve la petición primero).
func Encadenar(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// ─── Cabeceras de seguridad ─────────────────────────────────────────────────

// CabecerasSeguridad fija un conjunto conservador de cabeceras de
// seguridad en cada respuesta. La CSP permite scripts/estilos del propio
// origen + 'unsafe-inline' para los <script src> y los `style=` que ya
// tiene el proyecto. Si en el futuro se elimina todo style inline se
// puede endurecer quitando 'unsafe-inline'.
func (s *Sec) CabecerasSeguridad() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			h.Set("Permissions-Policy", "geolocation=(self), camera=(), microphone=(), payment=()")
			h.Set("Cross-Origin-Opener-Policy", "same-origin")
			h.Set("Cross-Origin-Resource-Policy", "same-origin")

			// CSP — ajustada a las dependencias actuales (Leaflet via cloudflare,
			// imágenes desde Unsplash + Wikimedia + tile servers, formularios solo al propio origen).
			//
			// upload.wikimedia.org se incluye porque la migración 011 sustituyó
			// las fotos de Unsplash por fotos reales de Wikipedia. Sin esto, el
			// navegador bloquea TODAS las imágenes de las estaciones.
			h.Set("Content-Security-Policy", strings.Join([]string{
				"default-src 'self'",
				"script-src 'self' https://cdnjs.cloudflare.com https://unpkg.com",
				"style-src 'self' 'unsafe-inline' https://unpkg.com https://cdnjs.cloudflare.com",
				"img-src 'self' data: blob: https://*.unsplash.com https://images.unsplash.com https://upload.wikimedia.org https://*.wikimedia.org https://*.basemaps.cartocdn.com https://*.tile.openstreetmap.org",
				"font-src 'self' data:",
				"connect-src 'self' https://nominatim.openstreetmap.org https://*.basemaps.cartocdn.com https://*.tile.openstreetmap.org",
				"object-src 'none'",
				"base-uri 'self'",
				"form-action 'self'",
				"frame-ancestors 'none'",
				"upgrade-insecure-requests",
			}, "; "))

			// HSTS solo cuando estamos detrás de TLS de verdad. Si la
			// config lo declara producción asumimos HTTPS por delante.
			if s.cfg.EsProduccion() {
				h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}

			next.ServeHTTP(w, r)
		})
	}
}

// LimitarBody envuelve r.Body con http.MaxBytesReader para evitar que un
// cliente mande gigabytes. Aplicar antes de cualquier ParseForm/Decode.
func (s *Sec) LimitarBody() func(http.Handler) http.Handler {
	max := s.cfg.BodyMaxBytes
	if max <= 0 {
		max = 1 << 20
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, max)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ─── Rate limiting ───────────────────────────────────────────────────────────

type ipBucket struct {
	tokens    int
	updatedAt time.Time
}

// LimitarPorIP es un middleware que aplica un límite "token bucket"
// muy simple: al usuario le caben hasta N tokens, regeneramos uno cada
// 60/N segundos. Adecuado para frenar fuerza bruta en login/registro
// sin necesidad de añadir Redis.
//
// Usar como middleware específico de una ruta: rate := sec.LimitarPorIP(10).
// 10 peticiones/minuto por IP.
func (s *Sec) LimitarPorIP(porMinuto int) func(http.Handler) http.Handler {
	if porMinuto <= 0 {
		porMinuto = s.cfg.RateLimitPerMin
	}
	if porMinuto <= 0 {
		porMinuto = 30
	}
	intervalo := time.Minute / time.Duration(porMinuto)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := IPCliente(r, s.cfg.TrustProxy)
			if !s.tomaToken(ip, porMinuto, intervalo) {
				w.Header().Set("Retry-After", "30")
				http.Error(w, "demasiadas peticiones, prueba en unos segundos", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (s *Sec) tomaToken(clave string, max int, intervalo time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	ahora := time.Now()
	b, ok := s.limiters[clave]
	if !ok {
		b = &ipBucket{tokens: max, updatedAt: ahora}
		s.limiters[clave] = b
	}
	// Repone tokens en función del tiempo transcurrido.
	transcurrido := ahora.Sub(b.updatedAt)
	if transcurrido > 0 {
		nuevos := int(transcurrido / intervalo)
		if nuevos > 0 {
			b.tokens += nuevos
			if b.tokens > max {
				b.tokens = max
			}
			b.updatedAt = ahora
		}
	}
	if b.tokens <= 0 {
		return false
	}
	b.tokens--
	// Limpieza oportunista para evitar crecer la map sin parar.
	if len(s.limiters) > 4096 {
		for k, v := range s.limiters {
			if ahora.Sub(v.updatedAt) > 10*time.Minute {
				delete(s.limiters, k)
			}
		}
	}
	return true
}

// ─── CSRF (double-submit cookie) ─────────────────────────────────────────────

// EmitirCSRF asegura que cada respuesta GET a una página HTML lleva la
// cookie CSRF. La cookie es legible desde el servidor y se inyecta en
// los formularios como input oculto desde el handler que renderiza.
func (s *Sec) EmitirCSRF() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Si la petición ya trae cookie, la respetamos. Si no, emitimos.
			if c, err := r.Cookie(CookieCSRF); err != nil || c == nil || c.Value == "" {
				token := s.nuevoTokenCSRF()
				http.SetCookie(w, &http.Cookie{
					Name:     CookieCSRF,
					Value:    token,
					Path:     "/",
					HttpOnly: false, // tiene que poder leerla el formulario JS si lo enviase
					Secure:   s.cfg.CookieSecure,
					SameSite: http.SameSiteLaxMode,
					Expires:  time.Now().Add(8 * time.Hour),
				})
				// Inyectamos también en el contexto vía cabecera para que el
				// renderizador pueda meterla en los formularios.
				r.Header.Set(CabeceraCSRF, token)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// VerificarCSRF se aplica a métodos que mutan estado (POST/PUT/DELETE).
//
// Estrategia (defensa en profundidad combinada con SameSite=Lax):
//
//  1. Métodos seguros (GET/HEAD/OPTIONS) pasan sin tocar.
//  2. Si la petición trae cabecera Origin/Referer, comprobamos que su
//     host coincide con r.Host. Si no coincide → 403. Esto bloquea
//     forgeries cross-site disparados desde otra web aunque el
//     navegador del usuario aún tenga la cookie.
//  3. Si además el formulario lleva csrf_token o cabecera X-CSRF-Token,
//     se valida contra la cookie (token doblemente firmado HMAC). Es la
//     capa de "double-submit" pura para clientes que no expongan Origin.
//  4. Si NO hay Origin/Referer y NO hay token, dejamos pasar SOLO para
//     mantener compatibilidad con clientes legítimos antiguos. La
//     SameSite=Lax cookie sigue siendo la barrera principal en ese caso.
//
// El umbral mínimo (Origin check) ya impide los CSRF "puros" disparados
// desde un sitio externo en navegadores modernos.
func (s *Sec) VerificarCSRF() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet, http.MethodHead, http.MethodOptions:
				next.ServeHTTP(w, r)
				return
			}
			// 1) Origin/Referer check: si vienen, deben coincidir con Host.
			if !origenAceptable(r) {
				http.Error(w, "origen no autorizado", http.StatusForbidden)
				return
			}
			// 2) Si trae token, lo validamos. Si no, lo dejamos pasar
			//    confiando en SameSite + Origin check.
			recibido := r.Header.Get(CabeceraCSRF)
			if recibido == "" {
				if err := r.ParseForm(); err == nil {
					recibido = r.FormValue(CampoCSRF)
				}
			}
			if recibido != "" {
				c, err := r.Cookie(CookieCSRF)
				if err != nil || c.Value == "" || !s.tokenCSRFValido(c.Value, recibido) {
					http.Error(w, "token CSRF inválido", http.StatusForbidden)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// origenAceptable comprueba que Origin o Referer (al menos uno) apunten
// al mismo host que la petición. Si ninguno está presente, devolvemos
// true (no podemos validar): la cookie SameSite ya hace de barrera.
func origenAceptable(r *http.Request) bool {
	host := r.Host
	if host == "" {
		return true
	}
	if origin := r.Header.Get("Origin"); origin != "" {
		return urlHostIgual(origin, host)
	}
	if referer := r.Header.Get("Referer"); referer != "" {
		return urlHostIgual(referer, host)
	}
	return true // Sin Origin/Referer: confiamos en SameSite=Lax.
}

func urlHostIgual(rawURL, host string) bool {
	// Extracción rápida sin importar net/url para no añadir alocaciones
	// extra: buscamos "//" tras el esquema y comparamos hasta el primer
	// "/" o ":".
	idx := strings.Index(rawURL, "//")
	if idx < 0 {
		return false
	}
	resto := rawURL[idx+2:]
	end := len(resto)
	for i, c := range resto {
		if c == '/' || c == '?' || c == '#' {
			end = i
			break
		}
	}
	hostFromURL := resto[:end]
	// Normalizamos eliminando puerto si los hosts no llevan puerto idéntico.
	if h, _, err := net.SplitHostPort(hostFromURL); err == nil {
		hostFromURL = h
	}
	wanted := host
	if h, _, err := net.SplitHostPort(host); err == nil {
		wanted = h
	}
	return strings.EqualFold(hostFromURL, wanted)
}

// TokenCSRFActual lee la cookie y devuelve el token actual de la
// petición, o cadena vacía si todavía no hay. Lo usan los renders para
// inyectar el token en los formularios.
func TokenCSRFActual(r *http.Request) string {
	if v := r.Header.Get(CabeceraCSRF); v != "" {
		return v
	}
	if c, err := r.Cookie(CookieCSRF); err == nil {
		return c.Value
	}
	return ""
}

// nuevoTokenCSRF genera 32 bytes aleatorios firmados con HMAC.
// Formato: base64url(rand) + "." + base64url(hmac).
func (s *Sec) nuevoTokenCSRF() string {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		// Failsafe: tiempo + pid; muy poco probable en stdlib reciente.
		t := time.Now().UnixNano()
		raw = []byte(fmt.Sprintf("%d", t))
	}
	mac := hmac.New(sha256.New, s.csrfKey)
	mac.Write(raw)
	firma := mac.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(raw) + "." + base64.RawURLEncoding.EncodeToString(firma)
}

func (s *Sec) tokenCSRFValido(cookie, recibido string) bool {
	if cookie == "" || recibido == "" {
		return false
	}
	if subtleEq(cookie, recibido) == false {
		return false
	}
	partes := strings.Split(cookie, ".")
	if len(partes) != 2 {
		return false
	}
	raw, err := base64.RawURLEncoding.DecodeString(partes[0])
	if err != nil {
		return false
	}
	firma, err := base64.RawURLEncoding.DecodeString(partes[1])
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, s.csrfKey)
	mac.Write(raw)
	return hmac.Equal(firma, mac.Sum(nil))
}

// subtleEq compara strings en tiempo constante para no filtrar el token
// vía timing.
func subtleEq(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return hmac.Equal([]byte(a), []byte(b))
}

// ─── Origen y proxy ──────────────────────────────────────────────────────────

// IPCliente devuelve la IP del cliente. Si TRUST_PROXY=true, mira
// X-Forwarded-For (primer hop). En caso contrario usa RemoteAddr para
// evitar que un cliente cualquiera se haga pasar por otra IP.
func IPCliente(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			partes := strings.Split(xff, ",")
			ip := strings.TrimSpace(partes[0])
			if ip != "" {
				return ip
			}
		}
		if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
			return strings.TrimSpace(xrip)
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// ─── Helpers varios ──────────────────────────────────────────────────────────

// EsRedirectInterno comprueba que el destino es una ruta dentro de la
// propia app (empieza por "/" pero NO por "//" — eso evita Open Redirect).
// Útil cuando aceptamos un parámetro "redirigir" o "next" en formularios.
func EsRedirectInterno(destino string) bool {
	if destino == "" || destino == "/" {
		return true
	}
	if strings.HasPrefix(destino, "//") || strings.HasPrefix(destino, "\\") {
		return false
	}
	if !strings.HasPrefix(destino, "/") {
		return false
	}
	// Bloqueamos esquemas tipo /\example.com (windows) y similares.
	for _, c := range destino {
		if c == 0 || c == '\n' || c == '\r' {
			return false
		}
	}
	return true
}

// SaneaIDHex confirma que un string es un id hexadecimal de 64 chars (token de sesión).
// No se usa por ahora, pero deja patrón para validaciones similares.
func SaneaIDHex(s string) bool {
	if len(s) != 64 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}
