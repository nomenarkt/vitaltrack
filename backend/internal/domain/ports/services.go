package ports

import (
	"time"

	"github.com/nomenarkt/vitaltrack/backend/internal/domain"
)

type AirtableService interface {
	FetchMedicines() ([]domain.Medicine, error)
	FetchStockEntries() ([]domain.StockEntry, error)
	FetchFinancialEntries(year int, month time.Month) ([]domain.FinancialEntry, error)
	UpdateMedicineLastAlertedDate(medicineID string, date time.Time) error
}

type TelegramService interface {
	SendTelegramMessage(text string) error
	PollForCommands(
		fetch func() ([]domain.Medicine, []domain.StockEntry, error),
		reportFn func(year, month int) (domain.MonthlyFinancialReport, error),
	)
}

type StockDataPort interface {
	FetchMedicines() ([]domain.Medicine, error)
	FetchStockEntries() ([]domain.StockEntry, error)
	FetchFinancialEntries(year int, month time.Month) ([]domain.FinancialEntry, error)
	CreateStockEntry(domain.StockEntry) error
	UpdateForecastDate(medicineID string, forecastDate, updatedAt time.Time) error
}

type FinancialDataPort interface {
	FetchFinancialEntries(year int, month time.Month) ([]domain.FinancialEntry, error)
}
