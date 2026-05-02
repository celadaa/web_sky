// Package main — punto de entrada del servidor SkiHub.
//
// Flujo de arranque:
//
//  1. Carga la configuración del entorno (.env + variables exportadas).
//  2. Conecta al PostgreSQL configurado.
//  3. Aplica las migraciones SQL pendientes (db/migrations/*.sql).
//  4. Asegura que existe al menos un usuario administrador (bcrypt).
//  5. Construye servicios + handlers y monta las rutas HTTP.
//  6. Escucha en el puerto configurado.
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"skihub/internal/config"
	"skihub/internal/db"
	"skihub/internal/handlers"
	"skihub/internal/repository"
	"skihub/internal/services"
)

func main() {
	log.SetPrefix("[SKIHUB] ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	bd, err := db.Conectar(cfg)
	if err != nil {
		log.Fatalf("conexión BD: %v", err)
	}
	defer bd.Close()

	bootCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := db.EjecutarMigraciones(bootCtx, bd, "db/migrations"); err != nil {
		log.Fatalf("migraciones: %v", err)
	}

	// Si todavía no hay admin, creamos uno con la contraseña indicada en
	// ADMIN_PASSWORD. Si no se ha definido, dejamos el sistema sin admin
	// y será necesario crearlo a mano por SQL.
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
			log.Printf("Administrador inicial creado: %s (cámbiale la contraseña al primer login)",
				cfg.AdminEmail)
		}
	} else {
		log.Println("ADMIN_PASSWORD vacía: no se creará admin automáticamente")
	}

	// Repositorios
	usuarioRepo  := repository.NuevoUsuarioRepo(bd)
	estacionRepo := repository.NuevoEstacionRepo(bd)
	favoritoRepo := repository.NuevoFavoritoRepo(bd)
	noticiaRepo  := repository.NuevoNoticiaRepo(bd)
	sesionRepo   := repository.NuevoSesionRepo(bd)
	pedidoRepo   := repository.NuevoPedidoRepo(bd)

	// Servicios
	app := &handlers.App{
		UsuarioSvc:  services.NuevoUsuarioService(usuarioRepo),
		EstacionSvc: services.NuevoEstacionService(estacionRepo, favoritoRepo),
		NoticiaSvc:  services.NuevoNoticiaService(noticiaRepo),
		SesionSvc:   services.NuevoSesionService(sesionRepo, usuarioRepo),
		FavoritoSvc: services.NuevoFavoritoService(favoritoRepo),
		PedidoSvc:   services.NuevoPedidoService(pedidoRepo, estacionRepo),
		// Servicio de pistas en directo (scraping cacheado de infonieve.es).
		NieveSvc: services.NuevoNieveService(),
	}

	plantillas, err := handlers.CargarPlantillas(cfg.AppTemplates)
	if err != nil {
		log.Fatalf("cargando plantillas: %v", err)
	}
	app.Plantillas = plantillas

	mux := http.NewServeMux()

	// Páginas
	mux.HandleFunc("/", app.Home)
	mux.HandleFunc("/estaciones", app.Estaciones)
	mux.HandleFunc("/estacion/", app.Estacion)
	mux.HandleFunc("/noticias", app.Noticias)
	mux.HandleFunc("/forfaits", app.Forfaits)
	mux.HandleFunc("/cesta", app.Cesta)
	mux.HandleFunc("/registro", app.Registro)
	mux.HandleFunc("/legal/aviso-legal", app.AvisoLegal)
	mux.HandleFunc("/legal/privacidad", app.PoliticaPrivacidad)
	mux.HandleFunc("/legal/cookies", app.PoliticaCookies)

	// Auth
	mux.HandleFunc("/login", app.Login)
	mux.HandleFunc("/logout", app.Logout)
	mux.HandleFunc("/cambiar-password", app.CambiarPassword)

	// Favoritos
	mux.HandleFunc("/favoritos", app.FavoritosPagina)
	mux.HandleFunc("/favorito/toggle", app.FavoritoToggle)

	// Admin
	mux.HandleFunc("/admin/usuarios", app.AdminUsuarios)
	mux.HandleFunc("/admin/usuario/", app.AdminUsuarioDetalle)
	mux.HandleFunc("/admin/usuarios/borrar", app.AdminBorrarUsuario)
	mux.HandleFunc("/admin/usuarios/reset", app.AdminResetPassword)
	mux.HandleFunc("/admin/usuarios/toggle-admin", app.AdminToggleAdmin)

	// API REST
	mux.HandleFunc("/api/usuarios", app.ApiUsuarios)
	mux.HandleFunc("/api/usuarios/", app.ApiUsuario)
	mux.HandleFunc("/api/estaciones", app.ApiEstaciones)
	mux.HandleFunc("/api/cesta/checkout", app.ApiCheckout)

	// Pistas en directo (scraping de infonieve.es, cacheado)
	mux.HandleFunc("/pistas", app.Pistas)
	mux.HandleFunc("/api/nieve/estaciones", app.ApiNieveEstaciones)
	mux.HandleFunc("/api/nieve/estaciones/", app.ApiNieveEstacion)
	mux.HandleFunc("/api/nieve/regiones", app.ApiNieveRegiones)

	// Estáticos
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(cfg.AppStatic))))

	servidor := &http.Server{
		Addr:              cfg.AppPort,
		Handler:           logMiddleware(mux),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("Servidor escuchando en http://localhost%s", cfg.AppPort)
	if err := servidor.ListenAndServe(); err != nil {
		log.Fatalf("servidor caído: %v", err)
	}
}

// logMiddleware registra cada petición. Útil tanto en desarrollo como
// en producción detrás de Nginx (Nginx ya pone access logs, pero éste
// nos da contexto de aplicación).
func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s desde %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
