package di_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/nomenarkt/vitaltrack/backend/internal/di"
	"github.com/nomenarkt/vitaltrack/backend/internal/domain"
	"github.com/nomenarkt/vitaltrack/backend/internal/logger"
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
				if err := os.Setenv("ENABLE_ALERT_TICKER", "true"); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.Unsetenv("ENABLE_ALERT_TICKER"); err != nil {
					t.Fatal(err)
				}
			}
			if tt.pollingEnabled {
				if err := os.Setenv("ENABLE_TELEGRAM_POLLING", "true"); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.Unsetenv("ENABLE_TELEGRAM_POLLING"); err != nil {
					t.Fatal(err)
				}
			}
			defer func() {
				if err := os.Unsetenv("ENABLE_ALERT_TICKER"); err != nil {
					t.Fatal(err)
				}
			}()
			defer func() {
				if err := os.Unsetenv("ENABLE_TELEGRAM_POLLING"); err != nil {
					t.Fatal(err)
				}
			}()

			tickerCalled := false
			pollingCalled := false
			origTicker := di.StartTickerFunc
			origPolling := di.PollingFunc
			di.StartTickerFunc = func(_ context.Context, _ di.Dependencies, _ time.Duration, _ func() time.Time) func() {
				tickerCalled = true
				return func() {}
			}
			di.PollingFunc = func(_ context.Context, _ di.Dependencies) { pollingCalled = true }
			defer func() {
				di.StartTickerFunc = origTicker
				di.PollingFunc = origPolling
			}()

			deps := di.Dependencies{Airtable: &envMockAirtable{}, Telegram: &envMockTelegram{}, Logger: logger.NewStdLogger()}
			di.StartFromEnv(context.Background(), deps)

			if tt.expectTicker != tickerCalled {
				t.Errorf("ticker call = %v, want %v", tickerCalled, tt.expectTicker)
			}
			if tt.expectPolling != pollingCalled {
				t.Errorf("polling call = %v, want %v", pollingCalled, tt.expectPolling)
			}
		})
	}
}

func TestNewApp(t *testing.T) {
	if err := os.Setenv("AIRTABLE_BASE_ID", "a"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("AIRTABLE_MEDICINES_TABLE", "b"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("AIRTABLE_ENTRIES_TABLE", "c"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("AIRTABLE_TOKEN", "d"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("TELEGRAM_BOT_TOKEN", "e"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("TELEGRAM_CHAT_ID", "f"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Unsetenv("AIRTABLE_BASE_ID"); err != nil {
			t.Fatal(err)
		}
		if err := os.Unsetenv("AIRTABLE_MEDICINES_TABLE"); err != nil {
			t.Fatal(err)
		}
		if err := os.Unsetenv("AIRTABLE_ENTRIES_TABLE"); err != nil {
			t.Fatal(err)
		}
		if err := os.Unsetenv("AIRTABLE_TOKEN"); err != nil {
			t.Fatal(err)
		}
		if err := os.Unsetenv("TELEGRAM_BOT_TOKEN"); err != nil {
			t.Fatal(err)
		}
		if err := os.Unsetenv("TELEGRAM_CHAT_ID"); err != nil {
			t.Fatal(err)
		}
	}()

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
				if err := os.Setenv("ENABLE_ALERT_TICKER", "true"); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.Unsetenv("ENABLE_ALERT_TICKER"); err != nil {
					t.Fatal(err)
				}
			}
			if tt.pollingEnabled {
				if err := os.Setenv("ENABLE_TELEGRAM_POLLING", "true"); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.Unsetenv("ENABLE_TELEGRAM_POLLING"); err != nil {
					t.Fatal(err)
				}
			}
			defer func() {
				if err := os.Unsetenv("ENABLE_ALERT_TICKER"); err != nil {
					t.Fatal(err)
				}
			}()
			defer func() {
				if err := os.Unsetenv("ENABLE_TELEGRAM_POLLING"); err != nil {
					t.Fatal(err)
				}
			}()

			tickerCalled := false
			pollingCalled := false
			origTicker := di.StartTickerFunc
			origPolling := di.PollingFunc
			di.StartTickerFunc = func(_ context.Context, _ di.Dependencies, _ time.Duration, _ func() time.Time) func() {
				tickerCalled = true
				return func() {}
			}
			di.PollingFunc = func(_ context.Context, _ di.Dependencies) { pollingCalled = true }
			defer func() {
				di.StartTickerFunc = origTicker
				di.PollingFunc = origPolling
			}()

			app := di.NewApp()
			if app == nil {
				t.Fatal("app is nil")
			}

			if tt.expectTicker != tickerCalled {
				t.Errorf("ticker call = %v, want %v", tickerCalled, tt.expectTicker)
			}
			if tt.expectPolling != pollingCalled {
				t.Errorf("polling call = %v, want %v", pollingCalled, tt.expectPolling)
			}
		})
	}
}
