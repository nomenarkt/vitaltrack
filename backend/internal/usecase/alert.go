package usecase

import (
	"fmt"
	"time"

	"github.com/nomenarkt/medicine-tracker/backend/internal/domain/ports"
	"github.com/nomenarkt/medicine-tracker/backend/internal/logic/forecast"
	"github.com/nomenarkt/medicine-tracker/backend/internal/logic/stockcalc"
)

const stockThreshold = 10.0

// StockChecker handles alerting when stock is low.
type StockChecker struct {
	Airtable ports.AirtableService
	Telegram ports.TelegramService
}

// CheckAndAlertLowStock scans medicines and alerts if below threshold.
func (s *StockChecker) CheckAndAlertLowStock() error {
	now := time.Now().UTC()

	meds, err := s.Airtable.FetchMedicines()
	if err != nil {
		return fmt.Errorf("fetch medicines failed: %w", err)
	}
	entries, err := s.Airtable.FetchStockEntries()
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
			_ = s.Telegram.SendTelegramMessage(alert)
		}
	}

	return nil
}

// OutOfStockService wraps forecast generation logic.
type OutOfStockService struct {
	Airtable ports.StockDataPort
}

// GenerateOutOfStockForecastMessage returns a summary of stock depletion.
func (s OutOfStockService) GenerateOutOfStockForecastMessage() (string, error) {
	meds, err := s.Airtable.FetchMedicines()
	if err != nil {
		return "", fmt.Errorf("fetch medicines failed: %w", err)
	}
	entries, err := s.Airtable.FetchStockEntries()
	if err != nil {
		return "", fmt.Errorf("fetch stock entries failed: %w", err)
	}

	return forecast.GenerateOutOfStockForecastMessage(meds, entries, time.Now().UTC(), s.Airtable), nil
}
