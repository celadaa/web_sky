// Package main — punto de entrada del servidor SkiHub.
//
// Flujo de arranque:
//
//  1. Carga la configuracion del entorno (.env + variables exportadas).
//  2. Conecta al PostgreSQL configurado.
//  3. Aplica las migraciones SQL pendientes (db/migrations/*.sql).
//  4. Asegura que existe al menos un usuario administrador (bcrypt).
//  5. Construye servicios + handlers y monta las rutas HTTP.
//  6. Encadena middlewares de seguridad (cabeceras, CSP, body limit, CSRF, rate limit).
//  7. Escucha en el puerto configurado.
package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"skihub/internal/config"
	"skihub/internal/db"
	"skihub/internal/handlers"
	"skihub/internal/repository"
	"skihub/internal/services"
)

// version se sobreescribe en build time con ldflags. Si no se setea queda "dev".
//
//	go build -ldflags "-X main.version=$(git rev-parse --short HEAD)" ./cmd/servidor
//
// Se expone en /healthz para saber que commit esta sirviendo.
var version = "dev"

func main() {
	log.SetPrefix("[SKIHUB] ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	bd, err := db.Conectar(cfg)
	if err != nil {
		log.Fatalf("conexion BD: %v", err)
	}
	defer bd.Close()

	// appCtx se cancela cuando el proceso recibe SIGINT/SIGTERM.
	// Lo usamos para detener la goroutine de sincronizacion de feeds.
	appCtx, stopApp := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stopApp()

	bootCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := db.EjecutarMigraciones(bootCtx, bd, "db/migrations"); err != nil {
		log.Fatalf("migraciones: %v", err)
	}

	// Si todavia no hay admin, creamos uno con la contrasena indicada en
	// ADMIN_PASSWORD. Si no se ha definido, dejamos el sistema sin admin.
	if cfg.AdminPassword != "" {
		hash, err := services.HashPassword(cfg.AdminPassword)
		if err != nil {
			log.Fatalf("hash admin: %v", err)
		}
		creado, err := db.AsegurarAdmin(bootCtx, bd, cfg.AdminEmail, cfg.AdminName, hash)
		if err != nil {
			log.Fatalf("asegurando admin: %v", err)
		}
		if creado {
			log.Printf("Administrador inicial creado: %s (cambia la contrasena al primer login)",
				cfg.AdminEmail)
		}
	} else {
		log.Println("ADMIN_PASSWORD vacia: no se creara admin automaticamente")
	}

	// Repositorios
	usuarioRepo := repository.NuevoUsuarioRepo(bd)
	estacionRepo := repository.NuevoEstacionRepo(bd)
	favoritoRepo := repository.NuevoFavoritoRepo(bd)
	noticiaRepo := repository.NuevoNoticiaRepo(bd)
	sesionRepo := repository.NuevoSesionRepo(bd)
	pedidoRepo := repository.NuevoPedidoRepo(bd)

	// Servicios + estado de seguridad.
	sec := handlers.NuevoSec(cfg)

	// Email — opcional. Si la inicializacion falla lo logueamos pero seguimos.
	emailSvc, err := services.NuevoEmailService(cfg, cfg.AppTemplates)
	if err != nil {
		log.Printf("AVISO: EmailService no inicializado: %v (los emails no se enviaran)", err)
		emailSvc = nil
	}

	// Google OAuth — opcional.
	googleAuth, err := services.NuevoGoogleOAuthService(bootCtx, cfg, usuarioRepo)
	if err != nil {
		log.Printf("AVISO: Google OAuth no inicializado: %v (login con Google desactivado)", err)
		googleAuth = nil
	}
	if googleAuth == nil && cfg.GoogleClientID == "" {
		log.Println("INFO: GOOGLE_CLIENT_ID vacio — login con Google desactivado")
	}

	// nieveSvc es el servicio de pistas en directo (scraping cacheado de infonieve.es).
	nieveSvc := services.NuevoNieveService()

	estacionSvc := services.NuevoEstacionService(estacionRepo, favoritoRepo)
	estacionSvc.NieveSvc = nieveSvc

	// feedSvc sincroniza noticias externas (RSS) en background cada 2 horas.
	feedSvc := services.NuevoFeedSyncService(noticiaRepo, cfg.AppStatic)

	app := &handlers.App{
		UsuarioSvc:  services.NuevoUsuarioService(usuarioRepo),
		EstacionSvc: estacionSvc,
		NoticiaSvc:  services.NuevoNoticiaService(noticiaRepo),
		FeedSvc:     feedSvc,
		SesionSvc:   services.NuevoSesionService(sesionRepo, usuarioRepo),
		FavoritoSvc: services.NuevoFavoritoService(favoritoRepo),
		PedidoSvc:   services.NuevoPedidoService(pedidoRepo, estacionRepo),
		NieveSvc:    nieveSvc,
		EmailSvc:    emailSvc,
		GoogleAuth:  googleAuth,
		Cfg:         cfg,
		Sec:         sec,
		BD:          bd,
		Version:     version,
	}

	// Sincronizacion automatica: ejecuta inmediatamente al arrancar y cada 2 horas.
	feedSvc.IniciarSincronizacionPeriodica(appCtx, 2*time.Hour)

	plantillas, err := handlers.CargarPlantillas(cfg.AppTemplates)
	if err != nil {
		log.Fatalf("cargando plantillas: %v", err)
	}
	app.Plantillas = plantillas

	mux := http.NewServeMux()

	// Rate-limiters especificos de endpoints sensibles.
	rlAuth := sec.LimitarPorIP(10)
	rlEscritura := sec.LimitarPorIP(20)

	// Health check
	mux.HandleFunc("/healthz", app.Healthz)

	// Paginas
	mux.HandleFunc("/", app.Home)
	mux.HandleFunc("/estaciones", app.Estaciones)
	mux.HandleFunc("/estacion/", app.Estacion)
	mux.HandleFunc("/noticias", app.Noticias)
	mux.HandleFunc("/forfaits", app.Forfaits)
	mux.HandleFunc("/cesta", app.Cesta)
	mux.HandleFunc("/pago", app.Pago)
	mux.HandleFunc("/planificar-estancia", app.PlanificarEstancia)
	mux.Handle("/registro", rlAuth(http.HandlerFunc(app.Registro)))
	mux.HandleFunc("/legal/aviso-legal", app.AvisoLegal)
	mux.HandleFunc("/legal/privacidad", app.PoliticaPrivacidad)
	mux.HandleFunc("/legal/cookies", app.PoliticaCookies)

	// Auth (rate-limited)
	mux.Handle("/login", rlAuth(http.HandlerFunc(app.Login)))
	mux.HandleFunc("/logout", app.Logout)
	mux.Handle("/cambiar-password", rlAuth(http.HandlerFunc(app.CambiarPassword)))

	// Login con Google (OAuth + OIDC) + verificacion de email.
	mux.Handle("/auth/google", rlAuth(http.HandlerFunc(app.AuthGoogleInicio)))
	mux.Handle("/auth/google/callback", rlAuth(http.HandlerFunc(app.AuthGoogleCallback)))
	mux.Handle("/confirmar-email", rlAuth(http.HandlerFunc(app.ConfirmarEmail)))

	// Favoritos
	mux.HandleFunc("/favoritos", app.FavoritosPagina)
	mux.Handle("/favorito/toggle", rlEscritura(http.HandlerFunc(app.FavoritoToggle)))

	// Admin
	mux.HandleFunc("/admin/usuarios", app.AdminUsuarios)
	mux.HandleFunc("/admin/usuario/", app.AdminUsuarioDetalle)
	mux.Handle("/admin/usuarios/borrar", rlEscritura(http.HandlerFunc(app.AdminBorrarUsuario)))
	mux.Handle("/admin/usuarios/reset", rlEscritura(http.HandlerFunc(app.AdminResetPassword)))
	mux.Handle("/admin/usuarios/toggle-admin", rlEscritura(http.HandlerFunc(app.AdminToggleAdmin)))

	// API REST
	mux.HandleFunc("/api/usuarios", app.ApiUsuarios)
	mux.HandleFunc("/api/usuarios/", app.ApiUsuario)
	mux.HandleFunc("/api/estaciones", app.ApiEstaciones)
	mux.Handle("/api/estacion/", rlEscritura(http.HandlerFunc(app.ApiParteEstacion)))
	mux.Handle("/api/cesta/checkout", rlEscritura(http.HandlerFunc(app.ApiCheckout)))

	// Pistas en directo (scraping de infonieve.es, cacheado)
	mux.HandleFunc("/pistas", app.Pistas)
	mux.HandleFunc("/api/nieve/estaciones", app.ApiNieveEstaciones)
	mux.HandleFunc("/api/nieve/estaciones/", app.ApiNieveEstacion)
	mux.HandleFunc("/api/nieve/regiones", app.ApiNieveRegiones)

	// Estaticos
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(cfg.AppStatic))))

	pila := handlers.Encadenar(mux,
		logMiddleware,
		sec.CabecerasSeguridad(),
		sec.LimitarBody(),
		sec.EmitirCSRF(),
		sec.VerificarCSRF(),
	)

	servidor := &http.Server{
		Addr:              cfg.AppPort,
		Handler:           pila,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 14,
	}

	log.Printf("Servidor escuchando en http://localhost%s (env=%s, cookieSecure=%v)",
		cfg.AppPort, cfg.AppEnv, cfg.CookieSecure)
	if err := servidor.ListenAndServe(); err != nil {
		log.Fatalf("servidor caido: %v", err)
	}
}

// logMiddleware registra cada peticion.
func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s desde %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
