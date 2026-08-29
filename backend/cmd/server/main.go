// Command server runs the sports-tracker backend: REST API, WebSocket stream,
// and the metric-processing pipeline.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sadovod04/sports-tracker/internal/config"
	"github.com/sadovod04/sports-tracker/internal/httpapi"
	"github.com/sadovod04/sports-tracker/internal/live"
	"github.com/sadovod04/sports-tracker/internal/store"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := waitForMigrations(ctx, cfg); err != nil {
		return err
	}

	st, err := store.New(ctx, cfg.DatabaseURL())
	if err != nil {
		return err
	}
	defer st.Close()

	hub := live.New()
	srv := httpapi.NewServer(cfg, st, hub)

	go reapStaleDevices(ctx, st)

	httpServer := &http.Server{
		Addr:              ":" + cfg.APIPort,
		Handler:           srv.Routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	case err := <-errCh:
		return err
	}

	shutCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	return httpServer.Shutdown(shutCtx)
}

// waitForMigrations retries the migration step so the container can start
// alongside Postgres in docker-compose.
func waitForMigrations(ctx context.Context, cfg config.Config) error {
	const attempts = 30
	for i := 1; i <= attempts; i++ {
		err := store.Migrate(ctx, cfg.DatabaseURL())
		if err == nil {
			slog.Info("migrations applied")
			return nil
		}
		slog.Warn("waiting for database", "attempt", i, "err", err)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return errors.New("database not reachable after retries")
}

// reapStaleDevices periodically flips silent trackers to offline.
func reapStaleDevices(ctx context.Context, st *store.Store) {
	t := time.NewTicker(30 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := st.MarkStaleOffline(context.Background(), 2*time.Minute); err != nil {
				slog.Warn("reap stale devices", "err", err)
			}
		}
	}
}
