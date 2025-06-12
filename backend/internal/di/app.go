package di

import (
	"context"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nomenarkt/vitaltrack/backend/internal/server"
)

var (
	// StartTickerFunc points to the ticker starter implementation.
	// Tests or callers should assign it to background.StartStockAlertTicker.
	StartTickerFunc func(context.Context, Dependencies, time.Duration, func() time.Time) func()

	// PollingFunc points to the Telegram polling starter implementation.
	// Tests or callers should assign it to StartTelegramPolling.
	PollingFunc func(context.Context, Dependencies)
)

// StartFromEnv starts optional background processes based on environment flags.
func StartFromEnv(ctx context.Context, deps Dependencies) {
	tickerInterval := 24 * time.Hour
	if val := os.Getenv("ALERT_TICKER_INTERVAL"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			tickerInterval = d
		}
	}
	if os.Getenv("ENABLE_ALERT_TICKER") == "true" && StartTickerFunc != nil {
		StartTickerFunc(ctx, deps, tickerInterval, time.Now)
	}
	if os.Getenv("ENABLE_TELEGRAM_POLLING") == "true" && PollingFunc != nil {
		PollingFunc(ctx, deps)
	}
}

// NewApp initializes the Fiber application with all routes and optional
// background processes. It resolves dependencies via Init() and returns the
// configured *fiber.App instance.
func NewApp() *fiber.App {
	app := fiber.New()

	deps := Init()

	server.SetupRoutes(app, deps.StockChecker, deps.ForecastSvc, deps.MedicineSvc, deps.Airtable, deps.Telegram)

	if PollingFunc == nil {
		PollingFunc = StartTelegramPolling
	}

	StartFromEnv(context.Background(), deps)

	return app
}
