package forecast_test

import (
	"testing"
	"time"

	"github.com/nomenarkt/vitaltrack/backend/internal/domain"
	"github.com/nomenarkt/vitaltrack/backend/internal/logic/forecast"
)

// ✅ Complete mock that satisfies StockDataPort
type mockStockDataPort struct{}

func (m *mockStockDataPort) FetchMedicines() ([]domain.Medicine, error) {
	return []domain.Medicine{
		{
			ID:           "med1",
			Name:         "Paracetamol",
			InitialStock: 10,
			DailyDose:    1,
			StartDate:    domain.NewFlexibleDate(time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)),
			UnitPerBox:   10,
		},
	}, nil
}

func (m *mockStockDataPort) FetchStockEntries() ([]domain.StockEntry, error) {
	return []domain.StockEntry{}, nil
}

func (m *mockStockDataPort) CreateStockEntry(entry domain.StockEntry) error {
	return nil
}

// ✅ Corrected method signature here
func (m *mockStockDataPort) UpdateForecastDate(medicineID string, startDate time.Time, outOfStockDate time.Time) error {
	return nil
}
func (m *mockStockDataPort) FetchFinancialEntries(int, time.Month) ([]domain.FinancialEntry, error) {
	return nil, nil
}

func TestGenerateOutOfStockForecastMessage(t *testing.T) {
	mock := &mockStockDataPort{}
	now := time.Date(2025, 6, 4, 0, 0, 0, 0, time.UTC)

	meds, err := mock.FetchMedicines()
	if err != nil {
		t.Fatalf("fetch meds: %v", err)
	}
	entries, err := mock.FetchStockEntries()
	if err != nil {
		t.Fatalf("fetch entries: %v", err)
	}

	msg := forecast.GenerateOutOfStockForecastMessage(meds, entries, now, mock)
	if msg == "" {
		t.Error("Expected non-empty forecast message")
	}
}
