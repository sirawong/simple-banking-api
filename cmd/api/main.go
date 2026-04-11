// @title           Banking API
// @version         1.0
// @description     A simple banking API with accounts and transactions
// @host            localhost:8080
// @BasePath        /

package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirawong/simple-banking-api/cmd/api/di"
	"github.com/sirawong/simple-banking-api/pkg/logger"
)

func main() {
	log := logger.ProvideGlobalLogger()

	app, cleanup, err := di.InitializeApp(log)
	if err != nil {
		log.Error("failed to initialize app", err)
		os.Exit(1)
	}
	defer cleanup()

	go func() {
		if err := app.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.Stop(ctx); err != nil {
		log.Warn("server forced to shutdown", "error", err)
	}
	log.Info("server stopped")
}
