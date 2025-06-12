package di

import (
	"context"
	"os"
	"time"
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
