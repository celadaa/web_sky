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
	log.SetPrefix("[SKIHUB] ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	dirPlantillas := getenv("SKIHUB_TEMPLATES", "web/templates")
	dirEstaticos := getenv("SKIHUB_STATIC", "web/static")
	rutaBD := getenv("SKIHUB_DB", "data/skihub.db")
	puerto := getenv("SKIHUB_PORT", ":8080")
	adminEmail := getenv("SKIHUB_ADMIN_EMAIL", "admin@skihub.local")
	adminNombre := getenv("SKIHUB_ADMIN_NOMBRE", "Administrador")
	adminPwd := getenv("SKIHUB_ADMIN_PASSWORD", "admin1234")

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

	hash, err := services.HashPassword(adminPwd)
	if err != nil {
		log.Fatalf("hash admin: %v", err)
	}
	creado, err := db.AsegurarAdmin(bd, adminEmail, adminNombre, hash)
	if err != nil {
		log.Fatalf("asegurando admin: %v", err)
	}
	if creado {
		log.Println("Se ha creado un administrador por defecto:")
		log.Printf("  email: %s", adminEmail)
		log.Printf("  pass:  %s", adminPwd)
		log.Println("  Cámbiala al entrar por primera vez.")
	}

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

	plantillas, err := handlers.CargarPlantillas(dirPlantillas)
	if err != nil {
		log.Fatalf("cargando plantillas: %v", err)
	}
	app.Plantillas = plantillas

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.Home)
	mux.HandleFunc("/estaciones", app.Estaciones)
	mux.HandleFunc("/estacion/", app.Estacion)
	mux.HandleFunc("/noticias", app.Noticias)
	mux.HandleFunc("/registro", app.Registro)
	mux.HandleFunc("/legal/aviso-legal", app.AvisoLegal)
	mux.HandleFunc("/legal/privacidad", app.PoliticaPrivacidad)
	mux.HandleFunc("/legal/cookies", app.PoliticaCookies)

	mux.HandleFunc("/login", app.Login)
	mux.HandleFunc("/logout", app.Logout)
	mux.HandleFunc("/cambiar-password", app.CambiarPassword)

	mux.HandleFunc("/favoritos", app.FavoritosPagina)
	mux.HandleFunc("/favorito/toggle", app.FavoritoToggle)

	mux.HandleFunc("/admin/usuarios", app.AdminUsuarios)
	mux.HandleFunc("/admin/usuario/", app.AdminUsuarioDetalle)
	mux.HandleFunc("/admin/usuarios/borrar", app.AdminBorrarUsuario)
	mux.HandleFunc("/admin/usuarios/reset", app.AdminResetPassword)
	mux.HandleFunc("/admin/usuarios/toggle-admin", app.AdminToggleAdmin)

	mux.HandleFunc("/api/usuarios", app.ApiUsuarios)
	mux.HandleFunc("/api/usuarios/", app.ApiUsuario)
	mux.HandleFunc("/api/estaciones", app.ApiEstaciones)

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(dirEstaticos))))

	log.Printf("Servidor iniciado en http://localhost%s", puerto)
	if err := http.ListenAndServe(puerto, logMiddleware(mux)); err != nil {
		log.Fatalf("servidor caído: %v", err)
	}
}

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
