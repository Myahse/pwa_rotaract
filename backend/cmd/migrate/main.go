package main

import (
	"flag"
	"log"
	"os"

	"github.com/rotaract-civ/backend/internal/config"
	"github.com/rotaract-civ/backend/internal/database"
)

func main() {
	direction := flag.String("direction", "up", "migration direction: up or down")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	switch *direction {
	case "up":
		if err := database.RunMigrations(cfg.DatabaseURL); err != nil {
			log.Fatalf("migrate up: %v", err)
		}
		log.Println("migrations applied")
	case "down":
		log.Fatal("down migrations are not enabled yet")
	default:
		log.Fatalf("unsupported direction %q", *direction)
	}

	os.Exit(0)
}
