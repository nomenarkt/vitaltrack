// Package background schedules periodic tasks like alert tickers.
package background

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/nomenarkt/vitaltrack/backend/internal/di"
	"github.com/nomenarkt/vitaltrack/backend/internal/logic/stockcalc"
	"github.com/nomenarkt/vitaltrack/backend/internal/util"
)

// StartStockAlertTicker begins a goroutine that checks stock levels at the given
// interval and sends Telegram alerts when medicines are running low. The
// returned function stops the ticker.
func StartStockAlertTicker(ctx context.Context, deps di.Dependencies, interval time.Duration, nowFn func() time.Time) (stop func()) {
	deps.Logger.Info(ctx, "alert ticker started", "interval", interval)
	stopCh := make(chan struct{})
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-stopCh:
				return
			default:
			}
			now := nowFn().UTC()

			meds, err := deps.Airtable.FetchMedicines()
			if err != nil {
				deps.Logger.Error(ctx, "fetch medicines failed", "error", err)
				deps.Logger.Info(ctx, "alert ticker completed")
				time.Sleep(interval)
				continue
			}

			entries, err := deps.Airtable.FetchStockEntries()
			if err != nil {
				deps.Logger.Error(ctx, "fetch stock entries failed", "error", err)
				deps.Logger.Info(ctx, "alert ticker completed")
				time.Sleep(interval)
				continue
			}

			for _, m := range meds {
				if m.DailyDose <= 0 {
					continue
				}

				stock := stockcalc.CurrentStockAt(m, entries, now)
				if stock <= 0 {
					continue
				}

				forecast := stockcalc.OutOfStockDateAt(m, stock, now)
				daysLeft := int(math.Floor(forecast.Sub(now).Hours() / 24))
				if daysLeft <= 10 {
					msg := fmt.Sprintf(
						"⚠️ *Refill Alert* for *%s* – runs out on *%s*\n(%.2f pills left)",
						util.EscapeMarkdown(m.Name),
						forecast.Format("2006-01-02"),
						stock,
					)
					if err := deps.Telegram.SendTelegramMessage(msg); err != nil {
						deps.Logger.Error(ctx, "telegram send failed", "medicine_id", m.ID, "error", err)
					} else {
						deps.Logger.Info(ctx, "alert sent", "medicine_id", m.ID)
					}
				}
			}

			deps.Logger.Info(ctx, "alert ticker completed")

			select {
			case <-stopCh:
				return
			case <-ticker.C:
			}
		}
	}()
	return func() { close(stopCh) }
}

func init() {
	di.StartTickerFunc = StartStockAlertTicker
}
