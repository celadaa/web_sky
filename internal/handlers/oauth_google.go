// Package handlers — login con Google (OAuth + OIDC).
//
// Endpoints:
//
//	GET /auth/google           → redirige a accounts.google.com con state.
//	GET /auth/google/callback  → valida state, intercambia code, crea sesión.
//
// Errores: cualquier fallo de OAuth/OIDC vuelve a /login con ?mensaje=...
// para no revelar detalles internos. Los logs sí registran el motivo.
package handlers

import (
	"errors"
	"log"
	"net/http"
	"time"

	"skihub/internal/services"
)

// CookieOAuthState es la cookie efímera que guarda el `state` durante el
// redirect a Google. HttpOnly + SameSite=Lax. Caduca en 10 minutos: tiempo
// más que suficiente para que el usuario complete el consent screen.
const CookieOAuthState = "skihub_oauth_state"

// AuthGoogleInicio: genera state, lo guarda en cookie y redirige a Google.
func (a *App) AuthGoogleInicio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	if a.GoogleAuth == nil {
		http.Redirect(w, r, "/login?mensaje=Login+con+Google+no+est%C3%A1+configurado", http.StatusSeeOther)
		return
	}
	state, err := services.GenerarEstado()
	if err != nil {
		log.Printf("ERROR generando state OAuth: %v", err)
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     CookieOAuthState,
		Value:    state,
		Path:     "/auth/google",
		MaxAge:   600, // 10 min
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		Secure:   a.CookieSecure(),
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, a.GoogleAuth.AuthCodeURL(state), http.StatusSeeOther)
}

// AuthGoogleCallback: valida state, intercambia code, resuelve usuario,
// crea sesión y redirige al usuario a /estaciones.
func (a *App) AuthGoogleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	if a.GoogleAuth == nil {
		http.Redirect(w, r, "/login?mensaje=Login+con+Google+no+est%C3%A1+configurado", http.StatusSeeOther)
		return
	}
	ip := ""
	if a.Cfg != nil {
		ip = IPCliente(r, a.Cfg.TrustProxy)
	}

	// Google añade ?error=access_denied cuando el usuario cancela.
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		log.Printf("AUTH/Google: callback con error=%q ip=%s", errParam, ip)
		borrarCookieEstado(w, a) // SetCookie ANTES de http.Redirect (que flushea headers).
		http.Redirect(w, r, "/login?mensaje=Has+cancelado+el+inicio+con+Google", http.StatusSeeOther)
		return
	}

	state := r.URL.Query().Get("state")
	cookieState, _ := r.Cookie(CookieOAuthState)
	// Limpiar siempre, gane o pierda. Importante: ANTES de http.Redirect.
	borrarCookieEstado(w, a)
	if cookieState == nil || cookieState.Value == "" || state == "" || cookieState.Value != state {
		log.Printf("AUTH/Google: state inválido ip=%s", ip)
		http.Redirect(w, r, "/login?mensaje=No+se+pudo+verificar+el+inicio+con+Google", http.StatusSeeOther)
		return
	}

	code := r.URL.Query().Get("code")
	claims, err := a.GoogleAuth.IntercambiarYVerificar(r.Context(), code)
	if err != nil {
		log.Printf("AUTH/Google: intercambio fallido ip=%s err=%v", ip, err)
		http.Redirect(w, r, "/login?mensaje=No+se+pudo+verificar+tu+cuenta+de+Google", http.StatusSeeOther)
		return
	}

	usuario, creado, err := a.GoogleAuth.ResolverUsuarioDesdeGoogle(r.Context(), claims)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrGoogleAccountTakeover):
			log.Printf("AUTH/Google: vinculación bloqueada (email no verificado) email=%s ip=%s",
				maskEmail(claims.Email), ip)
			http.Redirect(w, r, "/login?mensaje=Ya+existe+una+cuenta+con+ese+email."+
				"+Inicia+sesi%C3%B3n+normal+y+verif%C3%ADcalo+antes+de+vincular+Google", http.StatusSeeOther)
			return
		default:
			log.Printf("AUTH/Google: error resolviendo usuario ip=%s err=%v", ip, err)
			http.Error(w, "error interno del servidor", http.StatusInternalServerError)
			return
		}
	}

	sesion, err := a.SesionSvc.Crear(r.Context(), usuario.ID)
	if err != nil {
		log.Printf("ERROR creando sesión usuario=%d: %v", usuario.ID, err)
		http.Error(w, "error interno del servidor", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     CookieSesion,
		Value:    sesion.Token,
		Path:     "/",
		Expires:  sesion.ExpiraEn,
		HttpOnly: true,
		Secure:   a.CookieSecure(),
		SameSite: http.SameSiteLaxMode,
	})

	// Bienvenida si el usuario es nuevo y Google verificó el email.
	if creado && usuario.EmailVerificado && a.EmailSvc != nil {
		go func(nombre, email string) {
			if err := a.EmailSvc.SendWelcomeEmail(r.Context(), nombre, email); err != nil {
				log.Printf("EMAIL bienvenida (google) fallo email=%s: %v", maskEmail(email), err)
			}
		}(usuario.Nombre, usuario.Email)
	}

	log.Printf("AUTH/Google: login OK usuario=%d creado=%v ip=%s", usuario.ID, creado, ip)

	// destino: prioriza ?redirigir interno, si no /estaciones.
	destino := r.URL.Query().Get("redirigir")
	if !EsRedirectInterno(destino) {
		destino = "/estaciones"
	}
	if destino == "" {
		destino = "/estaciones"
	}
	http.Redirect(w, r, destino, http.StatusSeeOther)
}

// borrarCookieEstado emite una cookie de borrado con los mismos atributos
// con que se creó (Path, Secure, SameSite). Sin esto el navegador no la
// reemplaza y queda pegada hasta su expiración natural.
func borrarCookieEstado(w http.ResponseWriter, a *App) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieOAuthState,
		Value:    "",
		Path:     "/auth/google",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   a.CookieSecure(),
		SameSite: http.SameSiteLaxMode,
	})
}
