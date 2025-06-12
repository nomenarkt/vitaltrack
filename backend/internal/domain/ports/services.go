// Package ports defines interfaces for external services.
package ports

import (
	"time"

	"github.com/nomenarkt/vitaltrack/backend/internal/domain"
)

// AirtableService defines operations required from the Airtable API client.
type AirtableService interface {
	FetchMedicines() ([]domain.Medicine, error)
	FetchStockEntries() ([]domain.StockEntry, error)
	FetchFinancialEntries(year int, month time.Month) ([]domain.FinancialEntry, error)
	UpdateMedicineLastAlertedDate(medicineID string, date time.Time) error
}

// TelegramService defines methods for interacting with Telegram.
type TelegramService interface {
	SendTelegramMessage(text string) error
	PollForCommands(
		fetch func() ([]domain.Medicine, []domain.StockEntry, error),
		reportFn func(year, month int) (domain.MonthlyFinancialReport, error),
	)
}

// StockDataPort is used by use cases to persist and retrieve stock data.
type StockDataPort interface {
	FetchMedicines() ([]domain.Medicine, error)
	FetchStockEntries() ([]domain.StockEntry, error)
	FetchFinancialEntries(year int, month time.Month) ([]domain.FinancialEntry, error)
	CreateStockEntry(domain.StockEntry) error
	UpdateForecastDate(medicineID string, forecastDate, updatedAt time.Time) error
}

// FinancialDataPort fetches financial entries for reporting.
type FinancialDataPort interface {
	FetchFinancialEntries(year int, month time.Month) ([]domain.FinancialEntry, error)
}
