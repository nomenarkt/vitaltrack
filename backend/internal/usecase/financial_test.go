package usecase_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/nomenarkt/medicine-tracker/backend/internal/domain"
	"github.com/nomenarkt/medicine-tracker/backend/internal/usecase"
)

type mockFinanceRepo struct {
	entries []domain.FinancialEntry
}

func (m mockFinanceRepo) FetchFinancialEntries(year int, month time.Month) ([]domain.FinancialEntry, error) {
	return m.entries, nil
}

func TestGenerateFinancialReport(t *testing.T) {
	d1 := time.Date(2025, 6, 5, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2025, 6, 6, 0, 0, 0, 0, time.UTC)
	entries := []domain.FinancialEntry{
		{Date: domain.NewFlexibleDate(d1), Need: "Med", Contributor: "Alice", Amount: 10},
		{Date: domain.NewFlexibleDate(d1), Need: "Med", Contributor: "Bob", Amount: 5},
		{Date: domain.NewFlexibleDate(d1), Need: "Food", Contributor: "Alice", Amount: 2},
		{Date: domain.NewFlexibleDate(d2), Need: "Med", Contributor: "Charlie", Amount: 1},
	}

	svc := usecase.FinancialReportService{Repo: mockFinanceRepo{entries: entries}}
	rep, err := svc.GenerateFinancialReport(2025, 6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedNeeds := []domain.NeedReportBlock{
		{
			Need: "2025-06-05 Food",
			Contributors: []domain.ContributorAmount{
				{Name: "Alice", Amount: 2},
				{Name: "Bob", Amount: 0},
				{Name: "Charlie", Amount: 0},
			},
			Total: 2,
		},
		{
			Need: "2025-06-05 Med",
			Contributors: []domain.ContributorAmount{
				{Name: "Alice", Amount: 10},
				{Name: "Bob", Amount: 5},
				{Name: "Charlie", Amount: 0},
			},
			Total: 15,
		},
		{
			Need: "2025-06-06 Med",
			Contributors: []domain.ContributorAmount{
				{Name: "Alice", Amount: 0},
				{Name: "Bob", Amount: 0},
				{Name: "Charlie", Amount: 1},
			},
			Total: 1,
		},
	}
	expectedContrib := []domain.ContributorAmount{
		{Name: "Alice", Amount: 12},
		{Name: "Bob", Amount: 5},
		{Name: "Charlie", Amount: 1},
	}

	if !reflect.DeepEqual(rep.Needs, expectedNeeds) {
		t.Errorf("needs = %#v, want %#v", rep.Needs, expectedNeeds)
	}
	if !reflect.DeepEqual(rep.Contributors, expectedContrib) {
		t.Errorf("contributors = %#v, want %#v", rep.Contributors, expectedContrib)
	}
	if rep.Total != 18 {
		t.Errorf("total = %v, want 18", rep.Total)
	}
}
