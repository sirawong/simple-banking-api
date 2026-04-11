// @title           Banking API
// @version         1.0
// @description     A simple banking API with accounts and transactions
// @host            localhost:8080
// @BasePath        /

package main

import (
	"flag"
	"log"

	"github.com/sirawong/simple-banking-api/cmd/api/di"
	adapterdb "github.com/sirawong/simple-banking-api/internal/adapter/postgres"
	"github.com/sirawong/simple-banking-api/internal/config"
)

func main() {
	migrate := flag.Bool("migrate", false, "Run database migration")
	flag.Parse()

	cfg, err := config.ProvideConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if *migrate {
		db, err := adapterdb.ProvideDB(cfg)
		if err != nil {
			log.Fatalf("failed to connect to database: %v", err)
		}
		if err := adapterdb.AutoMigrate(db); err != nil {
			log.Fatalf("failed to migrate: %v", err)
		}
		log.Println("migration completed")
		return
	}

	engine, err := di.InitializeApp(cfg)
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	port := cfg.App.Port
	if port == "" {
		port = "8080"
	}
	log.Printf("starting server on :%s", port)
	if err := engine.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
