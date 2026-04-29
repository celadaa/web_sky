package models

import "time"

// Sesion representa una sesión activa de un usuario identificada por un
// token aleatorio que se envía al navegador como cookie HttpOnly.
type Sesion struct {
	Token     string
	UsuarioID int64
	CreadaEn  time.Time
	ExpiraEn  time.Time
}

// Vigente devuelve true si la sesión no ha expirado todavía.
func (s Sesion) Vigente() bool {
	return time.Now().Before(s.ExpiraEn)
}
