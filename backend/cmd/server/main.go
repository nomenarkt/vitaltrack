package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/nomenarkt/medicine-tracker/backend/internal/background"
	"github.com/nomenarkt/medicine-tracker/backend/internal/infra/airtable"
	"github.com/nomenarkt/medicine-tracker/backend/internal/infra/telegram"
	"github.com/nomenarkt/medicine-tracker/backend/internal/server"
	"github.com/nomenarkt/medicine-tracker/backend/internal/usecase"
)

func main() {
	_ = godotenv.Load()
	app := fiber.New()

	// === Instantiate service implementations ===
	at := airtable.NewClient()
	tg := telegram.NewClient()

	// === Compose usecases ===
	checker := &usecase.StockChecker{
		Airtable: at,
		Telegram: tg,
	}
	forecast := usecase.OutOfStockService{
		Airtable: at,
	}

	// ✅ Setup all HTTP routes with DI
	server.SetupRoutes(app, checker, forecast, at, tg)

	// 🔄 Start background stock check (daily) if enabled
	if os.Getenv("ENABLE_ALERT_TICKER") == "true" {
		background.StartStockAlertTicker(telegram.HandleOutOfStockCommand)
	}

	// 🚀 Run server
	log.Fatal(app.Listen(":8787"))
}
