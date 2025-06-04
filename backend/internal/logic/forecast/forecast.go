package forecast

import (
	"fmt"
	"time"

	"github.com/nomenarkt/medicine-tracker/backend/internal/domain"
	"github.com/nomenarkt/medicine-tracker/backend/internal/domain/ports"
	"github.com/nomenarkt/medicine-tracker/backend/internal/logic/stockcalc"
)

func GenerateOutOfStockForecastMessage(
	meds []domain.Medicine,
	entries []domain.StockEntry,
	now time.Time,
	repo ports.StockDataPort,
) string {
	var rows []string

	for _, m := range meds {
		stock := stockcalc.CurrentStockAt(m, entries, now)
		if stock <= 0 || m.DailyDose == 0 {
			continue
		}

		forecastDate := stockcalc.OutOfStockDateAt(m, stock, now)

		shouldUpdate := true
		if m.ForecastOutOfStockDate != nil {
			saved := m.ForecastOutOfStockDate.UTC().Format("2006-01-02")
			computed := forecastDate.Format("2006-01-02")
			if saved == computed {
				shouldUpdate = false
			}
		}

		if shouldUpdate {
			err := repo.UpdateForecastDate(m.ID, forecastDate, now)
			if err != nil {
				fmt.Printf("❌ Failed to update forecast for %s: %v\n", m.Name, err)
			} else {
				fmt.Printf("🆗 Updated forecast for %s to %s\n", m.Name, forecastDate.Format("2006-01-02"))
			}
		}

		rows = append(rows, fmt.Sprintf("• %s → %s", m.Name, forecastDate.Format("2006-01-02")))
	}

	return "*Out-of-Stock Forecast*\n\n" + stringJoin(rows, "\n")
}

// Helper for strings.Join without import clutter
func stringJoin(lines []string, sep string) string {
	result := ""
	for i, l := range lines {
		if i > 0 {
			result += sep
		}
		result += l
	}
	return result
}
