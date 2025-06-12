package usecase_test

import (
	"errors"
	"testing"
	"time"

	"github.com/nomenarkt/vitaltrack/backend/internal/domain"
	"github.com/nomenarkt/vitaltrack/backend/internal/usecase"
)

type mockRepo struct {
	meds    []domain.Medicine
	entries []domain.StockEntry
}

func (m mockRepo) FetchMedicines() ([]domain.Medicine, error) {
	return m.meds, nil
}
func (m mockRepo) FetchStockEntries() ([]domain.StockEntry, error) {
	return m.entries, nil
}
func (m mockRepo) FetchFinancialEntries(int, time.Month) ([]domain.FinancialEntry, error) {
	return nil, nil
}
func (m mockRepo) CreateStockEntry(domain.StockEntry) error              { return nil }
func (m mockRepo) UpdateForecastDate(string, time.Time, time.Time) error { return nil }

func TestGetStockInfo(t *testing.T) {
	now := time.Date(2025, 6, 4, 0, 0, 0, 0, time.UTC)
	med := domain.Medicine{
		ID: "m1", Name: "Med1", StartDate: domain.NewFlexibleDate(time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)),
		InitialStock: 10, DailyDose: 1, UnitPerBox: 10,
	}
	tests := []struct {
		name      string
		repo      mockRepo
		wantStock float64
		wantDate  time.Time
		wantErr   bool
	}{
		{
			name:      "found",
			repo:      mockRepo{meds: []domain.Medicine{med}},
			wantStock: 7,
			wantDate:  time.Date(2025, 6, 11, 0, 0, 0, 0, time.UTC),
		},
		{
			name:      "withEntries",
			repo:      mockRepo{meds: []domain.Medicine{med}, entries: []domain.StockEntry{{MedicineID: "m1", Quantity: 1, Unit: "box", Date: domain.NewFlexibleDate(now)}}},
			wantStock: 17,
			wantDate:  time.Date(2025, 6, 21, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "not_found",
			repo:    mockRepo{meds: []domain.Medicine{}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := usecase.MedicineService{Repo: tt.repo}
			info, err := svc.GetStockInfo("m1", now)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				if !errors.Is(err, usecase.ErrMedicineNotFound) {
					t.Fatalf("wrong error: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if info.CurrentStock != tt.wantStock {
				t.Errorf("stock=%.2f want %.2f", info.CurrentStock, tt.wantStock)
			}
			if !info.OutOfStockDate.Equal(tt.wantDate) {
				t.Errorf("date=%s want %s", info.OutOfStockDate.Format("2006-01-02"), tt.wantDate.Format("2006-01-02"))
			}
		})
	}
}
