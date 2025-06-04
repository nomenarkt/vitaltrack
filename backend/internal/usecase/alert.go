package usecase

import (
	"fmt"
	"time"

	"github.com/nomenarkt/medicine-tracker/backend/internal/infra/airtable"
	"github.com/nomenarkt/medicine-tracker/backend/internal/infra/telegram"
	"github.com/nomenarkt/medicine-tracker/backend/internal/logic/forecast"
	"github.com/nomenarkt/medicine-tracker/backend/internal/logic/stockcalc"
)

const stockThreshold = 10.0

// CheckAndAlertLowStock scans medicines and alerts if below threshold
func CheckAndAlertLowStock() error {
	now := time.Now().UTC()

	meds, err := airtable.FetchMedicines()
	if err != nil {
		return fmt.Errorf("fetch medicines failed: %w", err)
	}
	entries, err := airtable.FetchStockEntries()
	if err != nil {
		return fmt.Errorf("fetch stock entries failed: %w", err)
	}

	for _, m := range meds {
		stock := stockcalc.CurrentStockAt(m, entries, now)
		if stock < stockThreshold {
			alert := fmt.Sprintf("⚠️ *%s* is low:\n%.2f pills left.\nRefill before %s.",
				m.Name,
				stock,
				stockcalc.OutOfStockDateAt(m, stock, now).Format("Jan 2"),
			)
			_ = telegram.SendTelegramMessage(alert)
		}
	}

	return nil
}

// OutOfStockService wraps forecast generator for Telegram/HTTP reuse
type OutOfStockService struct{}

func (OutOfStockService) GenerateOutOfStockForecastMessage() (string, error) {

	meds, err := airtable.FetchMedicines()
	if err != nil {
		return "", fmt.Errorf("fetch medicines failed: %w", err)
	}
	entries, err := airtable.FetchStockEntries()
	if err != nil {
		return "", fmt.Errorf("fetch stock entries failed: %w", err)
	}

	return forecast.GenerateOutOfStockForecastMessage(meds, entries), nil
}
