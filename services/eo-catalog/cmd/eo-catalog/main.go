// Command eo-catalog is the STAC catalog ConnectRPC service.
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"connectrpc.com/connect"

	"github.com/ppusapati/space/api/p9e/space/earthobs/v1/earthobsv1connect"
	"github.com/ppusapati/space/pkg/closer"
	"github.com/ppusapati/space/pkg/db"
	"github.com/ppusapati/space/pkg/httpserver"
	"github.com/ppusapati/space/pkg/middleware"
	"github.com/ppusapati/space/pkg/observability"
	"github.com/ppusapati/space/services/eo-catalog/internal/config"
	"github.com/ppusapati/space/services/eo-catalog/internal/handlers"
	"github.com/ppusapati/space/services/eo-catalog/internal/repository"
	"github.com/ppusapati/space/services/eo-catalog/internal/service"
)

func main() {
	if err := run(); err != nil {
		// Use stderr because the logger may not be up yet.
		_, _ = os.Stderr.Write([]byte("eo-catalog: " + err.Error() + "\n"))
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := observability.NewLogger(observability.LogConfig{
		Level: cfg.LogLevel, Service: cfg.ServiceName, Environment: cfg.Environment,
	})
	logger.Info("starting eo-catalog",
		"http_addr", cfg.HTTPAddr,
		"metrics_addr", cfg.MetricsAddr,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	pool, err := db.Open(ctx, db.Config{DSN: cfg.DSN})
	cancel()
	if err != nil {
		return err
	}

	var c closer.Closer
	c.Add("db", func(_ context.Context) error { pool.Close(); return nil })

	repo := repository.New(pool)
	svc := service.New(repo)
	handler := handlers.NewCatalogHandler(svc, []byte(cfg.CursorSecret))

	mux := http.NewServeMux()
	path, h := earthobsv1connect.NewCatalogServiceHandler(
		handler,
		connect.WithInterceptors(
			middleware.Recovery(logger),
			middleware.CorrelationAndTenant(),
			middleware.AccessLog(),
		),
	)
	mux.Handle(path, h)

	// /metrics + /healthz + /readyz on a separate listener.
	var ready atomic.Bool
	ready.Store(true)
	registry := observability.NewMetricsRegistry()
	metrics := observability.MetricsServer(cfg.MetricsAddr, registry, ready.Load)
	metricsErr := make(chan error, 1)
	go func() {
		err := metrics.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			metricsErr <- err
		}
		close(metricsErr)
	}()
	c.Add("metrics", metrics.Shutdown)

	serveErr := httpserver.Run(context.Background(), httpserver.Config{
		Addr:            cfg.HTTPAddr,
		Handler:         mux,
		ShutdownTimeout: cfg.ShutdownTimeout,
	})
	if serveErr != nil {
		logger.Error("http server stopped with error", "err", serveErr)
	}
	closeCtx, closeCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer closeCancel()
	if err := c.Run(closeCtx, cfg.ShutdownTimeout); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
		return err
	}
	if e, ok := <-metricsErr; ok && e != nil {
		return e
	}
	return serveErr
}
