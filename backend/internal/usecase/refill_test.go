package usecase_test

import (
	"testing"
	"time"

	"github.com/nomenarkt/vitaltrack/backend/internal/domain"
	"github.com/nomenarkt/vitaltrack/backend/internal/usecase"
)

// Mocks for Airtable and Telegram

type mockAirtableRefill struct {
	meds    []domain.Medicine
	entries []domain.StockEntry
}

func (m *mockAirtableRefill) FetchMedicines() ([]domain.Medicine, error)      { return m.meds, nil }
func (m *mockAirtableRefill) FetchStockEntries() ([]domain.StockEntry, error) { return m.entries, nil }
func (m *mockAirtableRefill) FetchFinancialEntries(int, time.Month) ([]domain.FinancialEntry, error) {
	return nil, nil
}
func (m *mockAirtableRefill) UpdateMedicineLastAlertedDate(string, time.Time) error { return nil }
func (m *mockAirtableRefill) CreateStockEntry(domain.StockEntry) error              { return nil }
func (m *mockAirtableRefill) UpdateForecastDate(string, time.Time, time.Time) error { return nil }

type mockTelegramRefill struct{ msgs []string }

func (m *mockTelegramRefill) SendTelegramMessage(msg string) error {
	m.msgs = append(m.msgs, msg)
	return nil
}
func (m *mockTelegramRefill) PollForCommands(func() ([]domain.Medicine, []domain.StockEntry, error), func(int, int) (domain.MonthlyFinancialReport, error)) {
}

func TestCheckAndAlertNewRefills(t *testing.T) {
	now := time.Now().UTC().Truncate(24 * time.Hour)

	med := domain.Medicine{ID: "m1", Name: "Med1", UnitPerBox: 28}
	twoEntries := []domain.StockEntry{
		{MedicineID: "m1", Quantity: 1, Unit: "box", Date: domain.NewFlexibleDate(now)},
		{MedicineID: "m1", Quantity: 10, Unit: "pill", Date: domain.NewFlexibleDate(now)},
	}

	tests := []struct {
		name        string
		meds        []domain.Medicine
		entries     []domain.StockEntry
		expectCount int
	}{
		{
			name:        "single_entry",
			meds:        []domain.Medicine{med},
			entries:     []domain.StockEntry{twoEntries[0]},
			expectCount: 1,
		},
		{
			name:        "two_entries_same_medicine",
			meds:        []domain.Medicine{med},
			entries:     twoEntries,
			expectCount: 2,
		},
		{
			name:        "zero_quantity",
			meds:        []domain.Medicine{med},
			entries:     []domain.StockEntry{{MedicineID: "m1", Quantity: 0, Unit: "box", Date: domain.NewFlexibleDate(now)}},
			expectCount: 0,
		},
		{
			name:        "missing_medicine_id",
			meds:        []domain.Medicine{med},
			entries:     []domain.StockEntry{{MedicineID: "", Quantity: 2, Unit: "box", Date: domain.NewFlexibleDate(now)}},
			expectCount: 0,
		},
		{
			name:        "already_alerted_today",
			meds:        []domain.Medicine{{ID: "m1", Name: "Med1", UnitPerBox: 28, LastAlertedDate: &domain.FlexibleDate{Time: now}}},
			entries:     []domain.StockEntry{{MedicineID: "m1", Quantity: 1, Unit: "box", Date: domain.NewFlexibleDate(now)}},
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			at := &mockAirtableRefill{meds: tt.meds, entries: tt.entries}
			tg := &mockTelegramRefill{}

			checker := usecase.StockChecker{Airtable: at, Telegram: tg}
			if err := checker.CheckAndAlertNewRefills(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(tg.msgs) != tt.expectCount {
				t.Errorf("expected %d messages, got %d", tt.expectCount, len(tg.msgs))
			}
		})
	}
}
