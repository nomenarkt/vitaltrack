package usecase

import (
	"fmt"
	"log"
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
	log.Println("📡 Starting CheckAndAlertLowStock...")

	meds, err := s.Airtable.FetchMedicines()
	if err != nil {
		return fmt.Errorf("fetch medicines failed: %w", err)
	}
	log.Printf("📋 Fetched %d medicines", len(meds))

	entries, err := s.Airtable.FetchStockEntries()
	if err != nil {
		return fmt.Errorf("fetch stock entries failed: %w", err)
	}
	log.Printf("📦 Fetched %d stock entries", len(entries))

	for _, m := range meds {
		stock := stockcalc.CurrentStockAt(m, entries, now)
		log.Printf("🧪 %s: %.2f pills left", m.Name, stock)

		if stock < stockThreshold {
			alert := fmt.Sprintf("⚠️ *%s* is low:\n%.2f pills left.\nRefill before %s.",
				m.Name,
				stock,
				stockcalc.OutOfStockDateAt(m, stock, now).Format("2006-01-02"),
			)

			log.Printf("📲 Sending alert for %s", m.Name)
			err := s.Telegram.SendTelegramMessage(alert)
			if err != nil {
				log.Printf("❌ Telegram send failed: %v", err)
			} else {
				log.Println("✅ Telegram message sent")
			}

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
