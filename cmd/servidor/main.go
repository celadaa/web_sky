// Servidor web de SkiHub.
//
// Arquitectura por capas (Tema 3):
//
//	HTTP  →  handlers  →  services  →  repository  →  SQLite
//
// Punto de entrada único. Inicializa la BD, carga plantillas, registra rutas
// y arranca el servidor HTTP en el puerto 8080.
package main

import (
	"log"
	"net/http"
	"os"

	"skihub/internal/db"
	"skihub/internal/handlers"
	"skihub/internal/repository"
	"skihub/internal/services"
)

func main() {
	// Configuración del logger (Tema 3, p. 79): prefijo y metadatos.
	log.SetPrefix("[SKIHUB] ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Parámetros configurables por variable de entorno (con valores por defecto).
	dirPlantillas := getenv("SKIHUB_TEMPLATES", "web/templates")
	dirEstaticos := getenv("SKIHUB_STATIC", "web/static")
	rutaBD := getenv("SKIHUB_DB", "data/skihub.db")
	puerto := getenv("SKIHUB_PORT", ":8080")
	adminEmail := getenv("SKIHUB_ADMIN_EMAIL", "admin@skihub.local")
	adminNombre := getenv("SKIHUB_ADMIN_NOMBRE", "Administrador")
	adminPwd := getenv("SKIHUB_ADMIN_PASSWORD", "admin1234")

	// 1. Base de datos: abrir, migrar y sembrar.
	bd, err := db.Abrir(rutaBD)
	if err != nil {
		log.Fatalf("no se pudo abrir la BD: %v", err)
	}
	defer bd.Close()
	if err := db.Migrar(bd); err != nil {
		log.Fatalf("migración: %v", err)
	}
	if err := db.Semillas(bd); err != nil {
		log.Fatalf("semillas: %v", err)
	}

	// Asegurar que hay al menos un administrador.
	hash, err := services.HashPassword(adminPwd)
	if err != nil {
		log.Fatalf("hash admin: %v", err)
	}
	creado, err := db.AsegurarAdmin(bd, adminEmail, adminNombre, hash)
	if err != nil {
		log.Fatalf("asegurando admin: %v", err)
	}
	if creado {
		log.Println("╔══════════════════════════════════════════════════════════════╗")
		log.Println("║  No había administradores — se ha creado uno por defecto:    ║")
		log.Printf("║  email: %-52s║", adminEmail)
		log.Printf("║  pass:  %-52s║", adminPwd)
		log.Println("║  ⚠  Cámbiala al entrar por primera vez.                      ║")
		log.Println("╚══════════════════════════════════════════════════════════════╝")
	}

	// 2. Cadena de capas: repositorios -> servicios -> aplicación HTTP.
	usuarioRepo := repository.NuevoUsuarioRepo(bd)
	favoritoRepo := repository.NuevoFavoritoRepo(bd)
	sesionRepo := repository.NuevoSesionRepo(bd)

	app := &handlers.App{
		UsuarioSvc:  services.NuevoUsuarioService(usuarioRepo),
		EstacionSvc: services.NuevoEstacionService(repository.NuevoEstacionRepo(bd), favoritoRepo),
		NoticiaSvc:  services.NuevoNoticiaService(repository.NuevoNoticiaRepo(bd)),
		SesionSvc:   services.NuevoSesionService(sesionRepo, usuarioRepo),
		FavoritoSvc: services.NuevoFavoritoService(favoritoRepo),
	}

	// 3. Plantillas (cache al arranque).
	plantillas, err := handlers.CargarPlantillas(dirPlantillas)
	if err != nil {
		log.Fatalf("cargando plantillas: %v", err)
	}
	app.Plantillas = plantillas

	// 4. Rutas.
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.Home)
	mux.HandleFunc("/estaciones", app.Estaciones)
	mux.HandleFunc("/estacion/", app.Estacion)
	mux.HandleFunc("/noticias", app.Noticias)
	mux.HandleFunc("/registro", app.Registro)

	// Autenticación.
	mux.HandleFunc("/login", app.Login)
	mux.HandleFunc("/logout", app.Logout)
	mux.HandleFunc("/cambiar-password", app.CambiarPassword)

	// Favoritos (requieren sesión).
	mux.HandleFunc("/favoritos", app.FavoritosPagina)
	mux.HandleFunc("/favorito/toggle", app.FavoritoToggle)

	// Panel de administración (requiere rol admin).
	mux.HandleFunc("/admin/usuarios", app.AdminUsuarios)
	mux.HandleFunc("/admin/usuario/", app.AdminUsuarioDetalle)
	mux.HandleFunc("/admin/usuarios/borrar", app.AdminBorrarUsuario)
	mux.HandleFunc("/admin/usuarios/reset", app.AdminResetPassword)
	mux.HandleFunc("/admin/usuarios/toggle-admin", app.AdminToggleAdmin)

	// API REST de usuarios (PEC 3, Tema 4).
	//   • /api/usuarios   → colección: GET (listar), POST (crear).
	//   • /api/usuarios/  → recurso  : GET, PUT, DELETE con /{id}.
	// La autenticación de admin para POST/PUT/DELETE se aplica dentro
	// de los propios handlers (ApiUsuarios y ApiUsuario), igual que las
	// rutas /admin/* hacen con requerirAdmin.
	mux.HandleFunc("/api/usuarios", app.ApiUsuarios)
	mux.HandleFunc("/api/usuarios/", app.ApiUsuario)

	// API REST de estaciones (vista mapa). Solo lectura, pública.
	mux.HandleFunc("/api/estaciones", app.ApiEstaciones)

	// 5. Archivos estáticos (CSS, imágenes…).
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(dirEstaticos))))

	// 6. Arranque.
	log.Printf("Servidor iniciado en http://localhost%s", puerto)
	log.Printf("  • Inicio:        http://localhost%s/", puerto)
	log.Printf("  • Registro:      http://localhost%s/registro", puerto)
	log.Printf("  • Iniciar ses.:  http://localhost%s/login", puerto)
	log.Printf("  • Favoritas:     http://localhost%s/favoritos", puerto)
	log.Printf("  • Admin:         http://localhost%s/admin/usuarios", puerto)
	log.Printf("  • API usuarios:  http://localhost%s/api/usuarios", puerto)
	log.Printf("  • API estac.:    http://localhost%s/api/estaciones", puerto)
	if err := http.ListenAndServe(puerto, logMiddleware(mux)); err != nil {
		log.Fatalf("servidor caído: %v", err)
	}
}

// logMiddleware registra cada petición entrante (método, ruta, origen).
func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s desde %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func getenv(clave, defecto string) string {
	if v := os.Getenv(clave); v != "" {
		return v
	}
	return defecto
}
