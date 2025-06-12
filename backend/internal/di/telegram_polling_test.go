package di_test

import (
	"testing"
	"time"

	"github.com/nomenarkt/vitaltrack/backend/internal/di"
	"github.com/nomenarkt/vitaltrack/backend/internal/domain"
	"github.com/nomenarkt/vitaltrack/backend/internal/usecase"
)

type mockAirtable struct {
	medsCalled    bool
	entriesCalled bool
}

func (m *mockAirtable) FetchMedicines() ([]domain.Medicine, error) {
	m.medsCalled = true
	return []domain.Medicine{{ID: "1"}}, nil
}
func (m *mockAirtable) FetchStockEntries() ([]domain.StockEntry, error) {
	m.entriesCalled = true
	return []domain.StockEntry{}, nil
}
func (m *mockAirtable) CreateStockEntry(domain.StockEntry) error              { return nil }
func (m *mockAirtable) UpdateForecastDate(string, time.Time, time.Time) error { return nil }
func (m *mockAirtable) FetchFinancialEntries(int, time.Month) ([]domain.FinancialEntry, error) {
	return nil, nil
}
func (m *mockAirtable) UpdateMedicineLastAlertedDate(string, time.Time) error { return nil }

type mockFinanceRepo struct{ called bool }

func (m *mockFinanceRepo) FetchFinancialEntries(int, time.Month) ([]domain.FinancialEntry, error) {
	m.called = true
	return nil, nil
}

type mockTelegram struct{ done chan struct{} }

func (m *mockTelegram) SendTelegramMessage(string) error { return nil }
func (m *mockTelegram) PollForCommands(fetch func() ([]domain.Medicine, []domain.StockEntry, error), report func(int, int) (domain.MonthlyFinancialReport, error)) {
	fetch()
	report(2024, 6)
	close(m.done)
}

func TestStartTelegramPolling(t *testing.T) {
	at := &mockAirtable{}
	repo := &mockFinanceRepo{}
	tg := &mockTelegram{done: make(chan struct{})}
	deps := di.Dependencies{
		Airtable:     at,
		Telegram:     tg,
		FinancialSvc: usecase.FinancialReportService{Repo: repo},
	}

	di.StartTelegramPolling(deps)

	select {
	case <-tg.done:
	case <-time.After(time.Second):
		t.Fatal("polling not invoked")
	}

	if !at.medsCalled || !at.entriesCalled {
		t.Errorf("fetch functions not used")
	}
	if !repo.called {
		t.Errorf("financial report not called")
	}
}
