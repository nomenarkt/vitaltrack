// Command server starts the HTTP server.
package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/nomenarkt/vitaltrack/backend/internal/background"
	"github.com/nomenarkt/vitaltrack/backend/internal/di"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("godotenv load: %v", err)
	}

	di.StartTickerFunc = background.StartStockAlertTicker

	app := di.NewApp()

	if err := app.Listen(":8787"); err != nil {
		log.Printf("❌ Server failed to start: %v", err)
		os.Exit(1)
	}
}
