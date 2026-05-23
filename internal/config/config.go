// Package config carga la configuración de la aplicación a partir del entorno.
//
// Convenciones:
//   - Toda la configuración se lee al arrancar y queda en una struct Config.
//   - Las variables sensibles (contraseñas, etc.) NUNCA se hardcodean.
//   - Si falta una variable obligatoria, LoadFromEnv falla con un error claro
//     y el binario no llega a abrir un puerto.
//
// El fichero `.env` (en la raíz del proyecto) se carga automáticamente si
// existe; en producción basta con que las variables estén exportadas en
// el entorno (systemd, Docker, etc.) y el `.env` se omite.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config agrupa todos los parámetros de arranque del servidor.
type Config struct {
	// Entorno y endurecimiento
	AppEnv          string // "development" | "production"
	CookieSecure    bool   // marcar cookies como Secure (recomendado en producción)
	TrustProxy      bool   // honrar X-Forwarded-For si hay proxy delante
	BodyMaxBytes    int64  // límite global de cuerpo de petición
	RateLimitPerMin int    // peticiones por minuto por IP en endpoints sensibles
	CSRFSecret      string // semilla del token CSRF (32+ chars)

	// PostgreSQL
	DBHost            string
	DBPort            int
	DBUser            string
	DBPassword        string
	DBName            string
	DBSSLMode         string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration

	// HTTP
	AppPort      string
	AppTemplates string
	AppStatic    string

	// Admin inicial
	AdminEmail    string
	AdminName     string
	AdminPassword string

	// URL base pública (para construir enlaces en emails y redirects OAuth).
	AppBaseURL string

	// Google OAuth — opcional. Si ClientID/Secret están vacíos, el login con
	// Google queda desactivado y /auth/google devuelve 503.
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	// SMTP — opcional. Si SMTPHost está vacío o EmailDisabled=true, el
	// servicio de email loguea los enlaces por stdout en vez de enviar.
	SMTPHost      string
	SMTPPort      int
	SMTPUser      string
	SMTPPass      string
	SMTPFromEmail string
	SMTPFromName  string
	EmailDisabled bool
}

// EsProduccion devuelve true cuando APP_ENV=production.
func (c Config) EsProduccion() bool {
	return strings.EqualFold(c.AppEnv, "production") || strings.EqualFold(c.AppEnv, "prod")
}

// DSN devuelve la cadena de conexión libpq (key=value) que pgx acepta como
// driver `database/sql` registrado como "pgx". Mantenemos el formato libpq
// (y no la URL postgres://) porque es más legible en logs y menos propenso
// a errores con caracteres especiales en la contraseña.
func (c Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

// LoadFromEnv lee `.env` (si existe) y construye la Config validando
// las variables obligatorias.
//
// Variables obligatorias: DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, DB_SSLMODE.
// Variables opcionales con default: APP_PORT, APP_TEMPLATES, APP_STATIC,
// ADMIN_*, DB_MAX_OPEN_CONNS, DB_MAX_IDLE_CONNS, DB_CONN_MAX_LIFETIME_MIN,
// APP_ENV, COOKIE_SECURE, TRUST_PROXY, BODY_MAX_BYTES, RATE_LIMIT_PER_MIN, CSRF_SECRET.
func LoadFromEnv() (*Config, error) {
	// Cargamos .env si existe. Usamos Overload (no Load) para que los
	// valores del fichero PISEN cualquier variable que el usuario tenga
	// ya exportada en su shell. En desarrollo eso es lo que el usuario
	// espera (edita .env → cambia el comportamiento). En producción
	// normalmente no hay .env: las variables vienen de systemd / Docker
	// y el Overload no se aplica porque el fichero no existe.
	_ = godotenv.Overload()

	cfg := &Config{
		AppEnv:       getenv("APP_ENV", "development"),
		CSRFSecret:   getenv("CSRF_SECRET", ""),
		DBHost:       getenv("DB_HOST", ""),
		DBUser:       getenv("DB_USER", ""),
		DBPassword:   getenv("DB_PASSWORD", ""),
		DBName:       getenv("DB_NAME", ""),
		DBSSLMode:    getenv("DB_SSLMODE", ""),
		AppPort:      getenv("APP_PORT", ":8080"),
		AppTemplates: getenv("APP_TEMPLATES", "web/templates"),
		AppStatic:    getenv("APP_STATIC", "web/static"),

		AdminEmail:    getenv("ADMIN_EMAIL", "admin@skihub.local"),
		AdminName:     getenv("ADMIN_NAME", "Administrador"),
		AdminPassword: getenv("ADMIN_PASSWORD", ""),

		AppBaseURL: strings.TrimRight(getenv("APP_BASE_URL", "http://localhost:8080"), "/"),

		GoogleClientID:     getenv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getenv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getenv("GOOGLE_REDIRECT_URL", ""),

		SMTPHost:      getenv("SMTP_HOST", ""),
		SMTPUser:      getenv("SMTP_USER", ""),
		SMTPPass:      getenv("SMTP_PASS", ""),
		SMTPFromEmail: getenv("SMTP_FROM_EMAIL", ""),
		SMTPFromName:  getenv("SMTP_FROM_NAME", "Snowbreak"),
	}

	port, err := getenvInt("DB_PORT", 0)
	if err != nil {
		return nil, err
	}
	cfg.DBPort = port

	cfg.DBMaxOpenConns, err = getenvInt("DB_MAX_OPEN_CONNS", 25)
	if err != nil {
		return nil, err
	}
	cfg.DBMaxIdleConns, err = getenvInt("DB_MAX_IDLE_CONNS", 5)
	if err != nil {
		return nil, err
	}
	mins, err := getenvInt("DB_CONN_MAX_LIFETIME_MIN", 30)
	if err != nil {
		return nil, err
	}
	cfg.DBConnMaxLifetime = time.Duration(mins) * time.Minute

	bodyMax, err := getenvInt("BODY_MAX_BYTES", 1<<20) // 1 MiB
	if err != nil {
		return nil, err
	}
	cfg.BodyMaxBytes = int64(bodyMax)

	cfg.RateLimitPerMin, err = getenvInt("RATE_LIMIT_PER_MIN", 30)
	if err != nil {
		return nil, err
	}

	// SMTP_PORT — 587 (STARTTLS) por defecto. Si la variable está pero
	// no es un entero válido el binario no arranca.
	cfg.SMTPPort, err = getenvInt("SMTP_PORT", 587)
	if err != nil {
		return nil, err
	}

	// EMAIL_DISABLED y default sensato: si no se ha configurado SMTP_HOST
	// el envío queda implícitamente desactivado, da igual la variable.
	cfg.EmailDisabled = getenvBool("EMAIL_DISABLED", cfg.SMTPHost == "")

	// GoogleRedirectURL — si no se ha configurado explícitamente lo
	// derivamos de AppBaseURL para que en local con APP_BASE_URL=
	// http://localhost:8080 no haya que poner las dos variables.
	if cfg.GoogleRedirectURL == "" && cfg.AppBaseURL != "" {
		cfg.GoogleRedirectURL = cfg.AppBaseURL + "/auth/google/callback"
	}

	// COOKIE_SECURE: por defecto activo en producción, desactivado en dev
	// (HTTPS suele faltar en localhost y Secure haría que la cookie no
	// se envíe nunca).
	cfg.CookieSecure = getenvBool("COOKIE_SECURE", cfg.EsProduccion())
	cfg.TrustProxy = getenvBool("TRUST_PROXY", false)

	if err := cfg.validar(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// validar comprueba que las variables imprescindibles estén informadas y
// que los modos SSL sean válidos.
func (c *Config) validar() error {
	var faltan []string
	if c.DBHost == "" {
		faltan = append(faltan, "DB_HOST")
	}
	if c.DBPort == 0 {
		faltan = append(faltan, "DB_PORT")
	}
	if c.DBUser == "" {
		faltan = append(faltan, "DB_USER")
	}
	if c.DBPassword == "" {
		faltan = append(faltan, "DB_PASSWORD")
	}
	if c.DBName == "" {
		faltan = append(faltan, "DB_NAME")
	}
	if c.DBSSLMode == "" {
		faltan = append(faltan, "DB_SSLMODE")
	}
	if len(faltan) > 0 {
		return fmt.Errorf("faltan variables de entorno obligatorias: %s "+
			"(consulta .env.example)", strings.Join(faltan, ", "))
	}
	switch c.DBSSLMode {
	case "disable", "require", "verify-ca", "verify-full":
	default:
		return errors.New("DB_SSLMODE inválido: usa disable | require | verify-ca | verify-full")
	}
	if c.EsProduccion() {

		if c.DBSSLMode == "disable" && c.DBHost != "127.0.0.1" && c.DBHost != "localhost" && c.DBHost != "::1" {
			return errors.New("DB_SSLMODE=disable no es válido en producción salvo para PostgreSQL local: usa require | verify-ca | verify-full")
		}
		if c.AdminPassword != "" && len(c.AdminPassword) < 12 {
			return errors.New("ADMIN_PASSWORD demasiado corta para producción (mínimo 12 caracteres)")
		}
		if c.CSRFSecret == "" || len(c.CSRFSecret) < 32 {
			return errors.New("CSRF_SECRET requerida en producción (32+ caracteres aleatorios)")
		}
		// En producción el host base debe ser HTTPS — los enlaces de email
		// y los redirects OAuth pasarán por aquí.
		if !strings.HasPrefix(c.AppBaseURL, "https://") {
			return errors.New("APP_BASE_URL debe empezar por https:// en producción")
		}
	}

	// Coherencia de Google OAuth: si das uno tienes que dar los dos.
	hasID := c.GoogleClientID != ""
	hasSecret := c.GoogleClientSecret != ""
	if hasID != hasSecret {
		return errors.New("GOOGLE_CLIENT_ID y GOOGLE_CLIENT_SECRET deben configurarse juntos o ambos vacíos")
	}
	if hasID && c.GoogleRedirectURL == "" {
		return errors.New("GOOGLE_REDIRECT_URL no puede quedar vacío si GOOGLE_CLIENT_ID está configurado")
	}

	// Coherencia de SMTP: si EmailDisabled=false debe haber al menos host y from.
	if !c.EmailDisabled {
		if c.SMTPHost == "" || c.SMTPFromEmail == "" {
			return errors.New("SMTP_HOST y SMTP_FROM_EMAIL son obligatorios si EMAIL_DISABLED=false")
		}
		if c.SMTPPort <= 0 || c.SMTPPort > 65535 {
			return errors.New("SMTP_PORT fuera de rango (1..65535)")
		}
	}

	return nil
}

// getenv devuelve el valor de la variable o el defecto si está vacía.
func getenv(clave, defecto string) string {
	if v := os.Getenv(clave); v != "" {
		return v
	}
	return defecto
}

// getenvInt parsea un entero de la variable de entorno; si no está,
// devuelve el defecto. Si está pero no es entero, error explícito.
func getenvInt(clave string, defecto int) (int, error) {
	raw := os.Getenv(clave)
	if raw == "" {
		return defecto, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s debe ser un entero, recibido %q", clave, raw)
	}
	return n, nil
}

// getenvBool acepta 1/0, true/false, yes/no, on/off (case-insensitive).
func getenvBool(clave string, defecto bool) bool {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv(clave)))
	switch raw {
	case "":
		return defecto
	case "1", "true", "yes", "on", "y":
		return true
	case "0", "false", "no", "off", "n":
		return false
	default:
		return defecto
	}
}
