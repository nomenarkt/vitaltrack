package usecase_test

import (
	"testing"

	"github.com/nomenarkt/medicine-tracker/backend/internal/domain"
	"github.com/nomenarkt/medicine-tracker/backend/internal/usecase"
)

type mockAirtable struct {
	meds    []domain.Medicine
	entries []domain.StockEntry
}

func (m mockAirtable) FetchMedicines() ([]domain.Medicine, error) {
	return m.meds, nil
}
func (m mockAirtable) FetchStockEntries() ([]domain.StockEntry, error) {
	return m.entries, nil
}

type mockTelegram struct {
	sent []string
}

func (m *mockTelegram) SendTelegramMessage(msg string) error {
	m.sent = append(m.sent, msg)
	return nil
}

func TestCheckAndAlertLowStock(t *testing.T) {

	at := mockAirtable{
		meds: []domain.Medicine{
			{
				ID:           "low1",
				Name:         "LowMed",
				StartDate:    "2025-06-01",
				InitialStock: 10,
				DailyDose:    3,
				UnitPerBox:   10,
			},
		},
		entries: []domain.StockEntry{},
	}
	tg := &mockTelegram{}

	checker := usecase.StockChecker{
		Airtable: at,
		Telegram: tg,
	}

	err := checker.CheckAndAlertLowStock()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tg.sent) != 1 {
		t.Errorf("expected 1 telegram alert, got %d", len(tg.sent))
	}
}
