package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Healthz devuelve el estado del servicio en JSON. Lo usan:
//   - GitHub Actions tras un deploy, para validar que el binario nuevo levanta bien.
//   - Monitorización externa (UptimeRobot, BetterStack, Pingdom...).
//   - Tú mismo con `curl https://snowbreak.es/healthz` para un check rápido.
//
// Convenciones:
//   - 200 OK + JSON cuando todo va bien.
//   - 503 Service Unavailable + JSON cuando la BD no responde dentro del timeout.
//
// La cabecera Cache-Control evita que CDNs o proxys cacheen la respuesta.
func (a *App) Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")

	resp := map[string]string{
		"status":  "ok",
		"time":    time.Now().UTC().Format(time.RFC3339),
		"version": a.Version,
	}

	// Ping a la BD con timeout corto. Si falla devolvemos 503 para que
	// los balanceadores y el workflow de deploy lo detecten.
	if a.BD != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := a.BD.PingContext(ctx); err != nil {
			resp["status"] = "degraded"
			resp["db"] = "error"
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		resp["db"] = "ok"
	}

	_ = json.NewEncoder(w).Encode(resp)
}
