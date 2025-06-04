package stockcalc

import (
	"math"
	"time"

	"github.com/nomenarkt/medicine-tracker/backend/internal/domain"
)

// CurrentStockAt computes current pill stock based on:
// - Initial stock
// - Refill entries applied only on their specific date
// - Daily dose depletion from start date to now
func CurrentStockAt(m domain.Medicine, entries []domain.StockEntry, now time.Time) float64 {
	stock := m.InitialStock

	startDate, err := time.Parse("2006-01-02", m.StartDate)
	if err != nil {
		return stock // fallback to initial stock only
	}
	startDate = startDate.UTC()
	now = now.UTC()

	// Subtract consumed doses
	daysPassed := int(now.Sub(startDate).Hours() / 24)
	if daysPassed > 0 {
		stock -= float64(daysPassed) * m.DailyDose
	}

	// Apply refills that occur on today only (idempotent)
	for _, e := range entries {
		if e.MedicineID != m.ID {
			continue
		}
		if e.Date.Equal(now) {
			if e.Unit == "pill" {
				stock += float64(e.Quantity)
			} else {
				stock += float64(e.Quantity) * m.UnitPerBox
			}
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
