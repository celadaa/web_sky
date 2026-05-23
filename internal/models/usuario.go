// Package models define las entidades del dominio de Snowbreak.
// Siguiendo la arquitectura por capas del Tema 3, estas structs representan
// los datos que viajan entre repositorios, servicios y controladores.
package models

import "time"

// Usuario representa una cuenta de usuario registrada en Snowbreak.
// La contraseña se almacena SIEMPRE como hash, nunca en claro.
//
// Los campos con tipos puntero (*string, *time.Time) son NULLables en la
// tabla SQL — los repositorios los rellenan sólo cuando el SELECT los pide.
// Si una consulta antigua no los pide, su valor queda como nil sin romper
// el código previo.
type Usuario struct {
	ID            int64
	Nombre        string
	Email         string
	PasswordHash  string
	FechaRegistro time.Time
	EsAdmin       bool

	// Verificación de email (registro local).
	EmailVerificado   bool
	EmailVerificadoEn *time.Time

	// OAuth / OpenID Connect.
	GoogleID     *string // sub de Google. Único cuando no es nil.
	AvatarURL    *string // p.ej. picture de Google.
	AuthProvider string  // "local" | "google". Default "local".

	// Estado del token de verificación de email (sólo durante la ventana
	// de 24 h en la que el correo está pendiente de confirmarse).
	VerificationTokenHash      *string
	VerificationTokenExpiresAt *time.Time
}
