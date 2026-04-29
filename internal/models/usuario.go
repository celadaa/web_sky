// Package models define las entidades del dominio de Snowbreak.
// Siguiendo la arquitectura por capas del Tema 3, estas structs representan
// los datos que viajan entre repositorios, servicios y controladores.
package models

import "time"

// Usuario representa una cuenta de usuario registrada en Snowbreak.
// La contraseña se almacena SIEMPRE como hash, nunca en claro.
type Usuario struct {
	ID            int64
	Nombre        string
	Email         string
	PasswordHash  string
	FechaRegistro time.Time
	EsAdmin       bool
}
