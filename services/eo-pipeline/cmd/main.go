// Command eo-pipeline is the EO processing-job orchestration service.
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	pkgserver "p9e.in/samavaya/packages/connect/server"

	"github.com/ppusapati/space/services/eo-pipeline/api/eopipelinev1connect"
	"github.com/ppusapati/space/services/eo-pipeline/internal/config"
	"github.com/ppusapati/space/services/eo-pipeline/internal/handler"
	"github.com/ppusapati/space/services/eo-pipeline/internal/repository"
	"github.com/ppusapati/space/services/eo-pipeline/internal/services"
)

func main() {
	if err := run(); err != nil {
		_, _ = os.Stderr.Write([]byte("eo-pipeline: " + err.Error() + "\n"))
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	dbCtx, dbCancel := context.WithTimeout(context.Background(), 15*time.Second)
	pool, err := pgxpool.New(dbCtx, cfg.DSN)
	dbCancel()
	if err != nil {
		return err
	}
	defer pool.Close()

	repo := repository.New(pool)
	svc := services.New(repo)
	h, err := handler.NewPipelineHandler(svc)
	if err != nil {
		return err
	}

	mwCfg := pkgserver.MiddlewareConfig{
		EnableRecovery:       true,
		EnableRequestID:      true,
		EnableLogging:        true,
		EnableAuth:           false,
		EnableRLS:            false,
		EnableDB:             true,
		DBPool:               pool,
		SlowRequestThreshold: 5 * time.Second,
	}
	connectOption := pkgserver.NewConnectOption(mwCfg)

	mux := http.NewServeMux()
	path, hndl := eopipelinev1connect.NewPipelineServiceHandler(h, connectOption)
	mux.Handle(path, hndl)
	pkgserver.RegisterHealthEndpoints(mux)

	srvCfg := pkgserver.DefaultServerConfig(cfg.HTTPPort())
	srvCfg.AllowedOrigins = cfg.AllowedOrigins
	corsHandler := pkgserver.WrapWithCORS(mux, srvCfg.AllowedOrigins)
	finalHandler := pkgserver.WrapWithH2C(corsHandler)
	httpServer := pkgserver.NewHTTPServer(srvCfg, finalHandler)

	serveErr := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
		close(serveErr)
	}()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)
	select {
	case <-stopCh:
	case e, ok := <-serveErr:
		if ok {
			return e
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return err
	}
	if e, ok := <-serveErr; ok && e != nil {
		return e
	}
	return nil
}
