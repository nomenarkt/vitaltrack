package telegram

import (
	"github.com/nomenarkt/medicine-tracker/backend/internal/infra/airtable"
	"github.com/nomenarkt/medicine-tracker/backend/internal/logic/forecast"
)

func HandleOutOfStockCommand() error {
	meds, err := airtable.FetchMedicines()
	if err != nil {
		return err
	}
	entries, err := airtable.FetchStockEntries()
	if err != nil {
		return err
	}

	msg := forecast.GenerateOutOfStockForecastMessage(meds, entries)
	return SendTelegramMessage(msg)
}
