package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/nomenarkt/medicine-tracker/backend/internal/background"
	"github.com/nomenarkt/medicine-tracker/backend/internal/infra/telegram"
	"github.com/nomenarkt/medicine-tracker/backend/internal/server"
)

func main() {
	_ = godotenv.Load()
	app := fiber.New()

	// Setup all HTTP routes
	server.SetupRoutes(app)

	// 🔄 Start background stock check (daily) if enabled via env
	if os.Getenv("ENABLE_ALERT_TICKER") == "true" {
		background.StartStockAlertTicker(telegram.HandleOutOfStockCommand)
	}

	// ✅ Run server on correct port (8787)
	log.Fatal(app.Listen(":8787"))
}
