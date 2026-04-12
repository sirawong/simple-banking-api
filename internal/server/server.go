package server

import (
	"context"
	"net/http"

	"github.com/sirawong/simple-banking-api/internal/config"
	"github.com/sirawong/simple-banking-api/pkg/logger"
)

type App struct {
	httpServer *http.Server
}

func ProvideServer(cfg *config.Config, handler http.Handler) *App {
	port := cfg.App.Port
	if port == "" {
		port = "8080"
	}
	return &App{
		httpServer: &http.Server{
			Addr:    ":" + port,
			Handler: handler,
		},
	}
}

func (a *App) Start() error {
	logger.Info("server starting", "addr", a.httpServer.Addr)
	return a.httpServer.ListenAndServe()
}

func (a *App) Stop(ctx context.Context) error {
	return a.httpServer.Shutdown(ctx)
}
