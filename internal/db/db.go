// Package db gestiona la conexión a SQLite y la creación del esquema.
// Usamos el driver modernc.org/sqlite (Go puro, sin dependencias de C).
package db

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "modernc.org/sqlite"
)

// Abrir crea el fondo de conexiones a la base de datos indicada.
// El Ping fuerza una conexión real para detectar errores pronto.
func Abrir(ruta string) (*sql.DB, error) {
	bd, err := sql.Open("sqlite", ruta)
	if err != nil {
		return nil, fmt.Errorf("abriendo BD: %w", err)
	}
	if err := bd.Ping(); err != nil {
		return nil, fmt.Errorf("ping BD: %w", err)
	}
	log.Printf("BD SQLite abierta en %s", ruta)
	return bd, nil
}

// Migrar crea las tablas del esquema si no existen. Se ejecuta al
// iniciar el servidor, de forma idempotente.
//
// Para BBDD ya creadas se añaden columnas nuevas mediante ALTER TABLE
// capturando el error de "duplicate column" (SQLite no soporta IF NOT EXISTS
// sobre ADD COLUMN).
func Migrar(bd *sql.DB) error {
	sentencias := []string{
		`CREATE TABLE IF NOT EXISTS usuarios (
			id              INTEGER PRIMARY KEY AUTOINCREMENT,
			nombre          TEXT NOT NULL,
			email           TEXT NOT NULL UNIQUE,
			password_hash   TEXT NOT NULL,
			fecha_registro  DATETIME DEFAULT CURRENT_TIMESTAMP,
			es_admin        INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS estaciones (
			id               INTEGER PRIMARY KEY AUTOINCREMENT,
			nombre           TEXT NOT NULL,
			ubicacion        TEXT NOT NULL,
			distancia        REAL NOT NULL,
			temperatura      INTEGER NOT NULL,
			nieve_base       INTEGER NOT NULL,
			nieve_nueva      INTEGER NOT NULL,
			pistas_abiertas  INTEGER NOT NULL,
			pistas_totales   INTEGER NOT NULL,
			remontes_op      INTEGER NOT NULL,
			remontes_tot     INTEGER NOT NULL,
			ultima_nevada    TEXT NOT NULL,
			altitud          TEXT NOT NULL,
			km_esquiables    REAL NOT NULL,
			dificultad       TEXT NOT NULL,
			telefono         TEXT NOT NULL,
			imagen           TEXT NOT NULL,
			descripcion      TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS noticias (
			id              INTEGER PRIMARY KEY AUTOINCREMENT,
			titulo          TEXT NOT NULL,
			extracto        TEXT NOT NULL,
			categoria       TEXT NOT NULL,
			categoria_clase TEXT NOT NULL,
			fecha           DATE NOT NULL,
			imagen          TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS sesiones (
			token       TEXT PRIMARY KEY,
			usuario_id  INTEGER NOT NULL,
			creada_en   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			expira_en   DATETIME NOT NULL,
			FOREIGN KEY (usuario_id) REFERENCES usuarios(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS favoritos (
			usuario_id   INTEGER NOT NULL,
			estacion_id  INTEGER NOT NULL,
			creada_en    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (usuario_id, estacion_id),
			FOREIGN KEY (usuario_id)  REFERENCES usuarios(id)   ON DELETE CASCADE,
			FOREIGN KEY (estacion_id) REFERENCES estaciones(id) ON DELETE CASCADE
		)`,
	}
	for _, s := range sentencias {
		if _, err := bd.Exec(s); err != nil {
			return fmt.Errorf("migración fallida: %w", err)
		}
	}

	// Migración incremental para BBDD creadas antes de existir es_admin.
	if _, err := bd.Exec(`ALTER TABLE usuarios ADD COLUMN es_admin INTEGER NOT NULL DEFAULT 0`); err != nil {
		if !strings.Contains(err.Error(), "duplicate column") {
			return fmt.Errorf("alter usuarios es_admin: %w", err)
		}
	}

	log.Println("Esquema de BD actualizado")
	return nil
}

// Semillas inserta datos iniciales de estaciones y noticias si la BD
// está vacía, para que la web muestre contenido desde el primer arranque.
func Semillas(bd *sql.DB) error {
	var n int
	if err := bd.QueryRow(`SELECT COUNT(*) FROM estaciones`).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		if err := seedEstaciones(bd); err != nil {
			return err
		}
	}
	if err := bd.QueryRow(`SELECT COUNT(*) FROM noticias`).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		if err := seedNoticias(bd); err != nil {
			return err
		}
	}
	return nil
}

// AsegurarAdmin se llama al arrancar: si no existe ningún usuario con
// es_admin = 1, crea uno por defecto (o lo promociona si ya existe ese email).
// Pensado para que el operador pueda entrar siempre a /admin aunque la BD
// fuera creada en una versión anterior sin admins.
func AsegurarAdmin(bd *sql.DB, email, nombre, passwordHash string) (creado bool, err error) {
	var n int
	if err = bd.QueryRow(`SELECT COUNT(*) FROM usuarios WHERE es_admin = 1`).Scan(&n); err != nil {
		return false, err
	}
	if n > 0 {
		return false, nil
	}
	// ¿El email elegido ya existe como usuario normal? Entonces solo
	// lo promocionamos a admin para no duplicar la cuenta.
	var id int64
	err = bd.QueryRow(`SELECT id FROM usuarios WHERE email = ?`, email).Scan(&id)
	switch {
	case err == nil:
		_, err = bd.Exec(`UPDATE usuarios SET es_admin = 1 WHERE id = ?`, id)
		return err == nil, err
	case err == sql.ErrNoRows:
		_, err = bd.Exec(
			`INSERT INTO usuarios (nombre, email, password_hash, es_admin)
			 VALUES (?, ?, ?, 1)`, nombre, email, passwordHash,
		)
		return err == nil, err
	default:
		return false, err
	}
}

func seedEstaciones(bd *sql.DB) error {
	tipo := []struct {
		nombre                                   string
		ubicacion                                string
		distancia                                float64
		temp, nBase, nNueva, pAb, pTot, rOp, rTot int
		ultima, altitud                          string
		kmEsq                                    float64
		dificultad, telefono, imagen, descripcion string
	}{
		{"Candanchú", "Huesca, España", 368.8, -2, 120, 15, 42, 51, 22, 24,
			"Hace 4 días", "1.530m - 2.400m", 50.5,
			"Avanzado", "+34 974 373 194",
			"https://images.unsplash.com/photo-1478265409131-1f65c88f965c?q=80&w=1200&auto=format&fit=crop",
			"Candanchú es una estación de esquí pionera en España, conocida por su ambiente alpino y su exigente orografía. Con una cota máxima de 2.400 metros, ofrece algunas de las pistas más desafiantes del Pirineo, como el famoso Tubo de la Zapatilla. Además de sus pistas negras y rojas, cuenta con una excelente zona de debutantes protegida del viento, lo que la hace ideal para familias y principiantes."},
		{"Sierra Nevada", "Granada, España", 415.2, 1, 85, 5, 80, 131, 19, 21,
			"Hace 6 días", "2.100m - 3.300m", 110.0,
			"Medio", "+34 902 708 090",
			"https://images.unsplash.com/photo-1528048228650-0d2f4deedc67?q=80&w=1200&auto=format&fit=crop",
			"Sierra Nevada es la estación más al sur de Europa y la de mayor cota esquiable de España. Ofrece vistas impresionantes del Mediterráneo y una gran variedad de pistas para todos los niveles."},
		{"Grandvalira", "Andorra", 480.5, -4, 140, 25, 120, 138, 65, 70,
			"Hace 2 días", "1.710m - 2.640m", 210.0,
			"Todos los niveles", "+376 891 800",
			"https://images.unsplash.com/photo-1551524559-8af4e6624178?q=80&w=1200&auto=format&fit=crop",
			"Grandvalira es el mayor dominio esquiable de los Pirineos, con más de 200 km de pistas balizadas y modernas infraestructuras."},
		{"Formigal", "Huesca, España", 355.2, -3, 105, 10, 95, 147, 24, 28,
			"Hace 3 días", "1.500m - 2.250m", 182.0,
			"Medio", "+34 974 490 000",
			"https://images.unsplash.com/photo-1565992441121-4367c2967103?q=80&w=1200&auto=format&fit=crop",
			"Formigal ofrece cuatro valles conectados con gran variedad de pistas y una intensa vida de après-ski."},
		{"Baqueira Beret", "Lleida, España", 460.1, -5, 160, 20, 108, 167, 33, 36,
			"Hace 1 día", "1.500m - 2.610m", 167.0,
			"Avanzado", "+34 902 415 415",
			"https://images.unsplash.com/photo-1605540436563-5bca919ae766?q=80&w=1200&auto=format&fit=crop",
			"Baqueira Beret es la estación más grande del Pirineo catalán, famosa por la calidad de su nieve y sus amplias pistas."},
		{"Cerler", "Huesca, España", 402.7, -1, 90, 8, 58, 74, 17, 19,
			"Hace 5 días", "1.500m - 2.630m", 80.0,
			"Medio", "+34 974 551 111",
			"https://images.unsplash.com/photo-1724775640162-699696d70e6c?q=80&w=1200&auto=format&fit=crop",
			"Cerler es la estación de mayor altitud del Pirineo aragonés, ideal para esquiadores que buscan paisajes alpinos."},
	}
	stmt, err := bd.Prepare(`INSERT INTO estaciones
		(nombre, ubicacion, distancia, temperatura, nieve_base, nieve_nueva,
		 pistas_abiertas, pistas_totales, remontes_op, remontes_tot,
		 ultima_nevada, altitud, km_esquiables, dificultad, telefono,
		 imagen, descripcion)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, e := range tipo {
		if _, err := stmt.Exec(e.nombre, e.ubicacion, e.distancia, e.temp,
			e.nBase, e.nNueva, e.pAb, e.pTot, e.rOp, e.rTot,
			e.ultima, e.altitud, e.kmEsq, e.dificultad, e.telefono,
			e.imagen, e.descripcion); err != nil {
			return err
		}
	}
	log.Printf("Sembradas %d estaciones", len(tipo))
	return nil
}

func seedNoticias(bd *sql.DB) error {
	filas := []struct {
		titulo, extracto, cat, clase, fecha, img string
	}{
		{"Fuerte tormenta deja más de 50cm en Pirineos",
			"Las estaciones aragonesas y catalanas se preparan para un fin de semana espectacular tras las intensas precipitaciones...",
			"Nevada", "nevada", "2026-03-14",
			"https://images.unsplash.com/photo-1513342774453-5d76a9768b41?q=80&w=400&h=225&auto=format&fit=crop"},
		{"Cómo elegir tus botas de esquí ideales",
			"La comodidad es clave. Te damos los 5 aspectos fundamentales en los que fijarte antes de comprar o alquilar material...",
			"Consejos", "consejos", "2026-03-12",
			"https://images.unsplash.com/photo-1551698618-1dfe5d97d256?q=80&w=400&h=225&auto=format&fit=crop"},
		{"Campeonato Nacional de Slalom en Sierra Nevada",
			"Este fin de semana se reúnen los mejores atletas nacionales en la mítica pista del Río para disputarse el título...",
			"Evento", "evento", "2026-03-10",
			"https://plus.unsplash.com/premium_photo-1664302791901-52c6159eaf78?q=80&w=400&h=225&auto=format&fit=crop"},
		{"Nuevos forfaits conjuntos para la temporada",
			"Se anuncian los nuevos pases de temporada que permitirán esquiar en diferentes comunidades autónomas con descuentos...",
			"General", "general", "2026-03-08",
			"https://images.unsplash.com/photo-1486684338211-1a7ced564b0d?q=80&w=400&h=225&auto=format&fit=crop"},
		{"Alerta por riesgo de aludes nivel 4 en Andorra",
			"Tras las recientes precipitaciones y el aumento de las temperaturas, Protección Civil advierte del alto riesgo fuera de pistas...",
			"Nevada", "nevada", "2026-03-05",
			"https://images.unsplash.com/photo-1732692583018-2345548e4e5a?q=80&w=400&h=225&auto=format&fit=crop"},
		{"Preparación física pre-temporada: Evita lesiones",
			"Una buena rutina de fuerza y flexibilidad en piernas y core es esencial para disfrutar de la nieve de forma segura...",
			"Consejos", "consejos", "2026-03-01",
			"https://images.unsplash.com/photo-1596473536056-91eadf31189e?q=80&w=400&h=225&auto=format&fit=crop"},
	}
	stmt, err := bd.Prepare(`INSERT INTO noticias
		(titulo, extracto, categoria, categoria_clase, fecha, imagen)
		VALUES (?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, f := range filas {
		if _, err := stmt.Exec(f.titulo, f.extracto, f.cat, f.clase, f.fecha, f.img); err != nil {
			return err
		}
	}
	log.Printf("Sembradas %d noticias", len(filas))
	return nil
}
