package forecast

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/nomenarkt/medicine-tracker/backend/internal/domain"
	"github.com/nomenarkt/medicine-tracker/backend/internal/infra/airtable"
	"github.com/nomenarkt/medicine-tracker/backend/internal/logic/stockcalc"
)

// GenerateOutOfStockForecastMessage builds the daily forecast snapshot.
// It also updates the Airtable 'forecast_out_of_stock_date' field for each medicine.
func GenerateOutOfStockForecastMessage(
	meds []domain.Medicine,
	entries []domain.StockEntry,
) string {
	type row struct {
		Name  string
		Date  string
		Pills float64
	}

	now := time.Now().UTC()
	var rows []row

	for _, m := range meds {
		stock := stockcalc.CurrentStockAt(m, entries, now)
		if stock <= 0 || m.DailyDose == 0 {
			continue
		}
		forecastDate := stockcalc.OutOfStockDateAt(m, stock, now)

		// Compare to Airtable's saved forecast (if available)
		shouldUpdate := true
		if m.ForecastOutOfStockDate != nil {
			saved := m.ForecastOutOfStockDate.UTC().Format("2006-01-02")
			computed := forecastDate.Format("2006-01-02")
			if saved == computed {
				shouldUpdate = false
			}
		}

		if shouldUpdate {
			err := airtable.UpdateForecastDate(m.ID, forecastDate, now)
			if err != nil {
				log.Printf("❌ Failed to update forecast for %s: %v", m.Name, err)
			} else {
				log.Printf("🔄 Updated forecast for %s to %s", m.Name, forecastDate.Format("2006-01-02"))
			}
		}

		rows = append(rows, row{
			Name:  m.Name,
			Date:  forecastDate.Format("Jan 2, 2006"),
			Pills: stock,
		})
	}

	// Alphabetical sort
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Name < rows[j].Name
	})

	if len(rows) == 0 {
		return "✅ All medicines have enough stock."
	}

	message := "📦 *Out of Stock Forecasts*\n\n"
	for _, r := range rows {
		message += fmt.Sprintf("💊 %s → %s (%.2f pills left)\n", r.Name, r.Date, r.Pills)
	}
	return message
}
