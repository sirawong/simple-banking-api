package main

import (
	"os"

	"github.com/sirawong/simple-banking-api/cmd/migrate/di"
	"github.com/sirawong/simple-banking-api/pkg/logger"
)

func main() {
	log := logger.ProvideGlobalLogger()

	app, cleanup, err := di.InitializeMigrate(log)
	if err != nil {
		log.Error("failed to initialize", err)
		os.Exit(1)
	}
	defer cleanup()

	if err := app.Run(); err != nil {
		log.Error("migration failed", err)
		os.Exit(1)
	}

	log.Info("migration completed successfully")
}
