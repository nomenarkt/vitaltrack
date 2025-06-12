// Package forecast calculates medicine depletion dates.
package forecast

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/nomenarkt/vitaltrack/backend/internal/domain"
	"github.com/nomenarkt/vitaltrack/backend/internal/domain/ports"
	"github.com/nomenarkt/vitaltrack/backend/internal/logic/stockcalc"
)

// GenerateOutOfStockForecastMessage builds a Markdown-formatted forecast of when
// each medicine will run out of stock. It optionally updates the forecast date
// in the provided repository.
func GenerateOutOfStockForecastMessage(
	meds []domain.Medicine,
	entries []domain.StockEntry,
	now time.Time,
	repo ports.StockDataPort,
) string {
	type medicineForecast struct {
		Name         string
		ForecastDate time.Time
		ShouldUpdate bool
		ID           string
	}

	var forecasts []medicineForecast

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

		forecasts = append(forecasts, medicineForecast{
			Name:         m.Name,
			ForecastDate: forecastDate,
			ShouldUpdate: shouldUpdate,
			ID:           m.ID,
		})
	}

	// Sort by forecast date ascending
	sort.Slice(forecasts, func(i, j int) bool {
		return forecasts[i].ForecastDate.Before(forecasts[j].ForecastDate)
	})

	var rows []string

	for _, f := range forecasts {
		if f.ShouldUpdate && repo != nil {
			err := repo.UpdateForecastDate(f.ID, f.ForecastDate, now)
			if err != nil {
				log.Printf("❌ Failed to update forecast for %s: %v", f.Name, err)
			} else {
				log.Printf("🆗 Updated forecast for %s to %s", f.Name, f.ForecastDate.Format("2006-01-02"))
			}
		}
		rows = append(rows, fmt.Sprintf("%-22s → %s", f.Name, f.ForecastDate.Format("2006-01-02")))
	}

	return "*Out-of-Stock Forecast*\n\n```text\n" + stringJoin(rows, "\n") + "\n```"
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
