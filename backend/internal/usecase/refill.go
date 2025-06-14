package usecase

import (
	"fmt"
	"log"
	"time"

	"github.com/nomenarkt/vitaltrack/backend/internal/domain"
	"github.com/nomenarkt/vitaltrack/backend/internal/util"
)

// CheckAndAlertNewRefills notifies when new stock entries are recorded for today.
// It fetches medicines and stock entries, filters today's refills and sends a
// Telegram alert per entry.
func (s *StockChecker) CheckAndAlertNewRefills() error {
	now := time.Now().UTC()
	log.Printf("📡 Starting CheckAndAlertNewRefills...")

	meds, err := s.Airtable.FetchMedicines()
	if err != nil {
		return fmt.Errorf("fetch medicines failed: %w", err)
	}
	entries, err := s.Airtable.FetchStockEntries()
	if err != nil {
		return fmt.Errorf("fetch stock entries failed: %w", err)
	}

	medMap := make(map[string]domain.Medicine)
	for _, m := range meds {
		medMap[m.ID] = m
	}

	today := now.Format("2006-01-02")

	for _, e := range entries {
		if e.MedicineID == "" || e.Quantity <= 0 || e.Date.IsZero() {
			continue
		}
		if e.Date.UTC().Format("2006-01-02") != today {
			continue
		}

		med, ok := medMap[e.MedicineID]
		if !ok {
			continue
		}
		if med.LastAlertedDate != nil && med.LastAlertedDate.UTC().Format("2006-01-02") == today {
			continue
		}

		qty := e.Quantity
		if e.Unit == "box" {
			qty = qty * med.UnitPerBox
		}

		msg := fmt.Sprintf(
			"✅ Refill received: %s\n• Quantity: %.0f %s\n• Converted: %.0f pills\n• Date: %s",
			util.EscapeMarkdown(med.Name),
			e.Quantity,
			util.EscapeMarkdown(e.Unit),
			qty,
			e.Date.Format("2006-01-02"),
		)

		if err := s.Telegram.SendTelegramMessage(msg); err != nil {
			log.Printf("❌ Telegram send failed: %v", err)
		} else {
			log.Printf("✅ Refill alert sent: %s", med.Name)
		}
	}

	return nil
}
