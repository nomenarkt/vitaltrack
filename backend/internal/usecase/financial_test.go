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
	d3 := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		year    int
		month   int
		entries []domain.FinancialEntry
		want    domain.MonthlyFinancialReport
	}{
		{
			name:  "multiple_needs_same_date",
			year:  2025,
			month: 6,
			entries: []domain.FinancialEntry{
				{Date: domain.NewFlexibleDate(d1), Need: "Med", Contributor: "Alice", Amount: 10},
				{Date: domain.NewFlexibleDate(d1), Need: "Med", Contributor: "Bob", Amount: 5},
				{Date: domain.NewFlexibleDate(d1), Need: "Food", Contributor: "Alice", Amount: 2},
				{Date: domain.NewFlexibleDate(d2), Need: "Med", Contributor: "Charlie", Amount: 1},
			},
			want: domain.MonthlyFinancialReport{
				Year:  2025,
				Month: time.June,
				Needs: []domain.NeedReportBlock{
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
				},
				Contributors: []domain.ContributorAmount{
					{Name: "Alice", Amount: 12},
					{Name: "Bob", Amount: 5},
					{Name: "Charlie", Amount: 1},
				},
				Total: 18,
			},
		},
		{
			name:  "contributor_zero_amount",
			year:  2025,
			month: 7,
			entries: []domain.FinancialEntry{
				{Date: domain.NewFlexibleDate(d3), Need: "Med", Contributor: "Alice", Amount: 10},
				{Date: domain.NewFlexibleDate(d3), Need: "Med", Contributor: "Bob", Amount: 0},
			},
			want: domain.MonthlyFinancialReport{
				Year:  2025,
				Month: time.July,
				Needs: []domain.NeedReportBlock{
					{
						Need: "2025-07-01 Med",
						Contributors: []domain.ContributorAmount{
							{Name: "Alice", Amount: 10},
							{Name: "Bob", Amount: 0},
						},
						Total: 10,
					},
				},
				Contributors: []domain.ContributorAmount{
					{Name: "Alice", Amount: 10},
					{Name: "Bob", Amount: 0},
				},
				Total: 10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := usecase.FinancialReportService{Repo: mockFinanceRepo{entries: tt.entries}}
			rep, err := svc.GenerateFinancialReport(tt.year, tt.month)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(rep.Needs, tt.want.Needs) {
				t.Errorf("needs mismatch\nwant: %#v\n got: %#v", tt.want.Needs, rep.Needs)
			}
			if !reflect.DeepEqual(rep.Contributors, tt.want.Contributors) {
				t.Errorf("contributors mismatch\nwant: %#v\n got: %#v", tt.want.Contributors, rep.Contributors)
			}
			if rep.Total != tt.want.Total {
				t.Errorf("total = %v, want %v", rep.Total, tt.want.Total)
			}
		})
	}
}
