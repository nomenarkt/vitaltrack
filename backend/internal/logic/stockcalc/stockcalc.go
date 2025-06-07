package stockcalc

import (
	"math"
	"time"

	"github.com/nomenarkt/medicine-tracker/backend/internal/domain"
)

// CurrentStockAt computes current pill stock based on:
// - Initial stock
// - All past refill entries
// - Daily dose depletion from start date to now
func CurrentStockAt(m domain.Medicine, entries []domain.StockEntry, now time.Time) float64 {
	stock := m.InitialStock

	startDate := m.StartDate.Time // ✅ Use parsed time directly
	startDate = startDate.UTC()
	now = now.UTC()

	// Subtract consumed doses
	daysPassed := int(now.Truncate(24*time.Hour).Sub(startDate.Truncate(24*time.Hour)).Hours() / 24)
	if daysPassed > 0 {
		stock -= float64(daysPassed) * m.DailyDose
	}

	// Apply refills that occurred up to now (inclusive)
	for _, e := range entries {
		if e.MedicineID != m.ID {
			continue
		}
		if e.Date.IsZero() {
			continue // skip unparsed or missing date entries
		}
		if !e.Date.After(now) {
			qty := e.Quantity
			if e.Unit == "box" {
				qty *= m.UnitPerBox
			}
			stock += qty
		}
	}

	if stock < 0 {
		stock = 0
	}
	return math.Round(stock*100) / 100
}

// OutOfStockDateAt projects when the current stock will run out, assuming no future refills.
func OutOfStockDateAt(m domain.Medicine, stock float64, now time.Time) time.Time {
	if m.DailyDose == 0 {
		return now.AddDate(100, 0, 0) // effectively "never"
	}
	daysLeft := int(math.Floor(stock / m.DailyDose))
	return now.AddDate(0, 0, daysLeft)
}
