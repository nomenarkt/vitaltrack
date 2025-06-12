package di

import (
	"github.com/nomenarkt/vitaltrack/backend/internal/domain"
)

// StartTelegramPolling launches polling for Telegram bot commands.
func StartTelegramPolling(deps Dependencies) {
	go deps.Telegram.PollForCommands(
		func() ([]domain.Medicine, []domain.StockEntry, error) {
			meds, err := deps.Airtable.FetchMedicines()
			if err != nil {
				return nil, nil, err
			}
			entries, err := deps.Airtable.FetchStockEntries()
			if err != nil {
				return nil, nil, err
			}
			return meds, entries, nil
		},
		func(y, m int) (domain.MonthlyFinancialReport, error) {
			return deps.FinancialSvc.GenerateFinancialReport(y, m)
		},
	)
}
