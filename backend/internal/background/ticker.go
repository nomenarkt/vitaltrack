package background

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/nomenarkt/medicine-tracker/backend/internal/di"
	"github.com/nomenarkt/medicine-tracker/backend/internal/logic/stockcalc"
	"github.com/nomenarkt/medicine-tracker/backend/internal/util"
)

func StartStockAlertTicker(deps di.Dependencies, interval time.Duration, nowFn func() time.Time) (stop func()) {
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
				log.Println("❌ fetch medicines failed:", err)
				log.Println("🔁 Alert ticker completed")
				time.Sleep(interval)
				continue
			}

			entries, err := deps.Airtable.FetchStockEntries()
			if err != nil {
				log.Println("❌ fetch stock entries failed:", err)
				log.Println("🔁 Alert ticker completed")
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
						log.Println("❌ Telegram send failed:", err)
					} else {
						log.Printf("📣 Alert sent for %s", m.Name)
					}
				}
			}

			log.Println("🔁 Alert ticker completed")

			select {
			case <-stopCh:
				return
			case <-ticker.C:
			}
		}
	}()
	return func() { close(stopCh) }
}
