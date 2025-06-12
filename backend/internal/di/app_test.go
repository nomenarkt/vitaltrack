package di_test

import (
	"os"
	"testing"
	"time"

	"github.com/nomenarkt/vitaltrack/backend/internal/di"
	"github.com/nomenarkt/vitaltrack/backend/internal/domain"
)

type envMockAirtable struct{}

func (m *envMockAirtable) FetchMedicines() ([]domain.Medicine, error)      { return nil, nil }
func (m *envMockAirtable) FetchStockEntries() ([]domain.StockEntry, error) { return nil, nil }
func (m *envMockAirtable) FetchFinancialEntries(int, time.Month) ([]domain.FinancialEntry, error) {
	return nil, nil
}
func (m *envMockAirtable) CreateStockEntry(domain.StockEntry) error              { return nil }
func (m *envMockAirtable) UpdateForecastDate(string, time.Time, time.Time) error { return nil }
func (m *envMockAirtable) UpdateMedicineLastAlertedDate(string, time.Time) error { return nil }

type envMockTelegram struct{}

func (m *envMockTelegram) SendTelegramMessage(string) error { return nil }
func (m *envMockTelegram) PollForCommands(func() ([]domain.Medicine, []domain.StockEntry, error), func(int, int) (domain.MonthlyFinancialReport, error)) {
}

func TestStartFromEnv(t *testing.T) {
	tests := []struct {
		name           string
		tickerEnabled  bool
		pollingEnabled bool
		expectTicker   bool
		expectPolling  bool
	}{
		{name: "none"},
		{name: "ticker_only", tickerEnabled: true, expectTicker: true},
		{name: "polling_only", pollingEnabled: true, expectPolling: true},
		{name: "both", tickerEnabled: true, pollingEnabled: true, expectTicker: true, expectPolling: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.tickerEnabled {
				os.Setenv("ENABLE_ALERT_TICKER", "true")
			} else {
				os.Unsetenv("ENABLE_ALERT_TICKER")
			}
			if tt.pollingEnabled {
				os.Setenv("ENABLE_TELEGRAM_POLLING", "true")
			} else {
				os.Unsetenv("ENABLE_TELEGRAM_POLLING")
			}
			defer os.Unsetenv("ENABLE_ALERT_TICKER")
			defer os.Unsetenv("ENABLE_TELEGRAM_POLLING")

			tickerCalled := false
			pollingCalled := false
			origTicker := di.StartTickerFunc
			origPolling := di.PollingFunc
			di.StartTickerFunc = func(deps di.Dependencies, d time.Duration, nowFn func() time.Time) func() {
				tickerCalled = true
				return func() {}
			}
			di.PollingFunc = func(deps di.Dependencies) { pollingCalled = true }
			defer func() {
				di.StartTickerFunc = origTicker
				di.PollingFunc = origPolling
			}()

			deps := di.Dependencies{Airtable: &envMockAirtable{}, Telegram: &envMockTelegram{}}
			di.StartFromEnv(deps)

			if tt.expectTicker != tickerCalled {
				t.Errorf("ticker call = %v, want %v", tickerCalled, tt.expectTicker)
			}
			if tt.expectPolling != pollingCalled {
				t.Errorf("polling call = %v, want %v", pollingCalled, tt.expectPolling)
			}
		})
	}
}
