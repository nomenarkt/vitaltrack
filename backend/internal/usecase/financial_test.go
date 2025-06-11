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
	date := time.Date(2025, 6, 5, 0, 0, 0, 0, time.UTC)
	entries := []domain.FinancialEntry{
		{Date: domain.NewFlexibleDate(date), Need: "Med", Contributor: "Alice", Amount: 10},
		{Date: domain.NewFlexibleDate(date), Need: "Med", Contributor: "Bob", Amount: 5},
		{Date: domain.NewFlexibleDate(date), Need: "Food", Contributor: "Alice", Amount: 2},
	}

	svc := usecase.FinancialReportService{Repo: mockFinanceRepo{entries: entries}}
	rep, err := svc.GenerateFinancialReport(2025, 6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedNeeds := []domain.NeedReportBlock{
		{
			Need:         "Food",
			Contributors: []domain.ContributorAmount{{Name: "Alice", Amount: 2}},
			Total:        2,
		},
		{
			Need:         "Med",
			Contributors: []domain.ContributorAmount{{Name: "Alice", Amount: 10}, {Name: "Bob", Amount: 5}},
			Total:        15,
		},
	}
	expectedContrib := []domain.ContributorAmount{{Name: "Alice", Amount: 12}, {Name: "Bob", Amount: 5}}

	if !reflect.DeepEqual(rep.Needs, expectedNeeds) {
		t.Errorf("needs = %#v, want %#v", rep.Needs, expectedNeeds)
	}
	if !reflect.DeepEqual(rep.Contributors, expectedContrib) {
		t.Errorf("contributors = %#v, want %#v", rep.Contributors, expectedContrib)
	}
	if rep.Total != 17 {
		t.Errorf("total = %v, want 17", rep.Total)
	}
}
