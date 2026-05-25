package services

import (
	"context"
	"log"
	"sync"
	"time"

	"skihub/internal/feeds"
	"skihub/internal/repository"
)

// FeedSyncService orquesta la sincronización periódica de noticias externas.
//
// Flujo por ciclo:
//  1. Obtiene la lista de fuentes activas.
//  2. Lanza la descarga de cada fuente en paralelo (con timeout individual).
//  3. Para cada noticia obtenida llama a NoticiaRepo.UpsertExterno,
//     que inserta si es nueva o actualiza título/extracto/imagen si ya existe.
//  4. Registra errores por fuente pero no aborta el ciclo completo.
type FeedSyncService struct {
	Repo    *repository.NoticiaRepo
	Fetcher *feeds.Fetcher
	// Sources es la lista de fuentes. Si es nil se usan feeds.ActiveSources().
	Sources []feeds.Source
}

// NuevoFeedSyncService construye el servicio con los valores por defecto.
func NuevoFeedSyncService(repo *repository.NoticiaRepo, staticDir string) *FeedSyncService {
	return &FeedSyncService{
		Repo:    repo,
		Fetcher: feeds.NewFetcher(staticDir),
	}
}

// Sync ejecuta un ciclo completo de sincronización: descarga todas las
// fuentes activas y persiste las noticias en la base de datos.
// Es seguro llamarlo desde una goroutine de fondo.
func (s *FeedSyncService) Sync(ctx context.Context) {
	sources := s.Sources
	if len(sources) == 0 {
		sources = feeds.ActiveSources()
	}
	if len(sources) == 0 {
		log.Println("[feed_sync] no hay fuentes activas configuradas")
		return
	}

	log.Printf("[feed_sync] iniciando sincronización con %d fuente(s)", len(sources))
	inicio := time.Now()

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		totalNew int
		totalErr int
	)

	for _, src := range sources {
		src := src // captura para goroutine
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Timeout individual por fuente: no bloqueamos el ciclo si una fuente tarda.
			srcCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
			defer cancel()

			noticias, err := s.Fetcher.FetchSource(srcCtx, src)
			if err != nil {
				log.Printf("[feed_sync] ERROR fuente %q: %v", src.Name, err)
				mu.Lock()
				totalErr++
				mu.Unlock()
				return
			}

			nuevas := 0
			for _, n := range noticias {
				upserted, upsertErr := s.Repo.UpsertExterno(ctx, n)
				if upsertErr != nil {
					log.Printf("[feed_sync] upsert %q: %v", n.OriginalURL, upsertErr)
					continue
				}
				if upserted {
					nuevas++
				}
			}
			log.Printf("[feed_sync] fuente %q: %d noticias obtenidas, %d nuevas/actualizadas",
				src.Name, len(noticias), nuevas)
			mu.Lock()
			totalNew += nuevas
			mu.Unlock()
		}()
	}

	wg.Wait()
	log.Printf("[feed_sync] ciclo completado en %s — %d nuevas/actualizadas, %d fuentes con error",
		time.Since(inicio).Round(time.Millisecond), totalNew, totalErr)
}

// IniciarSincronizacionPeriodica arranca una goroutine que llama a Sync
// inmediatamente al arrancar y luego cada `intervalo`.
// El contexto ctx permite detener el bucle al apagar el servidor.
func (s *FeedSyncService) IniciarSincronizacionPeriodica(ctx context.Context, intervalo time.Duration) {
	go func() {
		log.Printf("[feed_sync] sincronización automática activa (intervalo: %s)", intervalo)

		// Primera sincronización al arrancar (con un pequeño retraso para que
		// el servidor esté completamente listo antes de hacer peticiones externas).
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
		s.Sync(ctx)

		ticker := time.NewTicker(intervalo)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("[feed_sync] deteniendo sincronización periódica")
				return
			case <-ticker.C:
				s.Sync(ctx)
			}
		}
	}()
}
