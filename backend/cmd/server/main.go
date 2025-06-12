package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/nomenarkt/vitaltrack/backend/internal/background"
	"github.com/nomenarkt/vitaltrack/backend/internal/di"
	"github.com/nomenarkt/vitaltrack/backend/internal/server"
)

func main() {
	_ = godotenv.Load()
	app := fiber.New()

	// ⛓️ Resolve dependencies via central initializer
	deps := di.Init()

	// ✅ Setup all HTTP routes with DI
	server.SetupRoutes(app, deps.StockChecker, deps.ForecastSvc, deps.Airtable, deps.Telegram)

	// 🔄 Start background stock check if enabled
	tickerInterval := 24 * time.Hour
	if val := os.Getenv("ALERT_TICKER_INTERVAL"); val != "" {
		d, err := time.ParseDuration(val)
		if err != nil {
			log.Printf("invalid ALERT_TICKER_INTERVAL %q: %v", val, err)
		} else {
			tickerInterval = d
		}
	}
	if os.Getenv("ENABLE_ALERT_TICKER") == "true" {
		background.StartStockAlertTicker(deps, tickerInterval, time.Now)
	}

	// 🧭 Start Telegram bot polling for `/stock` commands if enabled
	if os.Getenv("ENABLE_TELEGRAM_POLLING") == "true" {
		di.StartTelegramPolling(deps)
	}

	// 🚀 Run server
	log.Fatal(app.Listen(":8787"))
}
