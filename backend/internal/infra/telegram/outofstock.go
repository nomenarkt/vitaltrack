package telegram

import (
	"time"

	"github.com/nomenarkt/vitaltrack/backend/internal/infra/airtable"
	"github.com/nomenarkt/vitaltrack/backend/internal/logic/forecast"
)

// HandleOutOfStockCommand sends an out-of-stock forecast via Telegram.
func HandleOutOfStockCommand() error {
	at := airtable.NewClient()
	tg := NewClient()

	meds, err := at.FetchMedicines()
	if err != nil {
		return err
	}
	entries, err := at.FetchStockEntries()
	if err != nil {
		return err
	}

	msg := forecast.GenerateOutOfStockForecastMessage(meds, entries, time.Now().UTC(), airtable.NewClient())
	return tg.SendTelegramMessage(msg)
}
