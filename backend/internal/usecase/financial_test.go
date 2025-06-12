package usecase_test

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/nomenarkt/vitaltrack/backend/internal/domain"
	"github.com/nomenarkt/vitaltrack/backend/internal/usecase"
)

type mockFinanceRepo struct {
	entries []domain.FinancialEntry
}

func (m mockFinanceRepo) FetchFinancialEntries(year int, month time.Month) ([]domain.FinancialEntry, error) {
	return m.entries, nil
}

func sortContributors(input []domain.ContributorAmount, order []string) []domain.ContributorAmount {
	orderMap := map[string]int{}
	for i, name := range order {
		orderMap[name] = i
	}
	sort.SliceStable(input, func(i, j int) bool {
		return orderMap[input[i].Name] < orderMap[input[j].Name]
	})
	return input
}

func TestGenerateFinancialReport(t *testing.T) {
	d1 := time.Date(2025, 6, 5, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2025, 6, 6, 0, 0, 0, 0, time.UTC)
	d3 := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
	d4 := time.Date(2025, 8, 10, 0, 0, 0, 0, time.UTC)
	d5 := time.Date(2025, 9, 20, 0, 0, 0, 0, time.UTC)

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
				{Date: domain.NewFlexibleDate(d1), NeedLabel: "Med", NeedAmount: 20, AmountContributed: 10, Contributor: "Alice"},
				{Date: domain.NewFlexibleDate(d1), NeedLabel: "Med", NeedAmount: 20, AmountContributed: 5, Contributor: "Bob"},
				{Date: domain.NewFlexibleDate(d1), NeedLabel: "Food", NeedAmount: 5, AmountContributed: 2, Contributor: "Alice"},
				{Date: domain.NewFlexibleDate(d2), NeedLabel: "Med", NeedAmount: 8, AmountContributed: 1, Contributor: "Charlie"},
			},
			want: domain.MonthlyFinancialReport{
				Year:  2025,
				Month: time.June,
				Needs: []domain.NeedReportBlock{
					{
						Need:       "2025-06-05 Food",
						NeedAmount: 5,
						Contributors: []domain.ContributorAmount{
							{Name: "Alice", Amount: 2},
							{Name: "Bob", Amount: 0},
							{Name: "Charlie", Amount: 0},
						},
						Total: 2,
					},
					{
						Need:       "2025-06-05 Med",
						NeedAmount: 20,
						Contributors: []domain.ContributorAmount{
							{Name: "Alice", Amount: 10},
							{Name: "Bob", Amount: 5},
							{Name: "Charlie", Amount: 0},
						},
						Total: 15,
					},
					{
						Need:       "2025-06-06 Med",
						NeedAmount: 8,
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
				{Date: domain.NewFlexibleDate(d3), NeedLabel: "Med", NeedAmount: 10, AmountContributed: 10, Contributor: "Alice"},
				{Date: domain.NewFlexibleDate(d3), NeedLabel: "Med", NeedAmount: 10, AmountContributed: 0, Contributor: "Bob"},
			},
			want: domain.MonthlyFinancialReport{
				Year:  2025,
				Month: time.July,
				Needs: []domain.NeedReportBlock{
					{
						Need:       "2025-07-01 Med",
						NeedAmount: 10,
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
		{
			name:  "mixed_zero_and_positive",
			year:  2025,
			month: 8,
			entries: []domain.FinancialEntry{
				{Date: domain.NewFlexibleDate(d4), NeedLabel: "Food", NeedAmount: 15, AmountContributed: 0, Contributor: "Alice"},
				{Date: domain.NewFlexibleDate(d4), NeedLabel: "Food", NeedAmount: 15, AmountContributed: 5, Contributor: "Bob"},
				{Date: domain.NewFlexibleDate(d4), NeedLabel: "Food", NeedAmount: 15, AmountContributed: 0, Contributor: "Charlie"},
			},
			want: domain.MonthlyFinancialReport{
				Year:  2025,
				Month: time.August,
				Needs: []domain.NeedReportBlock{
					{
						Need:       "2025-08-10 Food",
						NeedAmount: 15,
						Contributors: []domain.ContributorAmount{
							{Name: "Alice", Amount: 0},
							{Name: "Bob", Amount: 5},
							{Name: "Charlie", Amount: 0},
						},
						Total: 5,
					},
				},
				Contributors: []domain.ContributorAmount{
					{Name: "Alice", Amount: 0},
					{Name: "Bob", Amount: 5},
					{Name: "Charlie", Amount: 0},
				},
				Total: 5,
			},
		},
		{
			name:  "all_zero_contributions",
			year:  2025,
			month: 9,
			entries: []domain.FinancialEntry{
				{Date: domain.NewFlexibleDate(d5), NeedLabel: "Med", NeedAmount: 100, AmountContributed: 0, Contributor: "Alice"},
				{Date: domain.NewFlexibleDate(d5), NeedLabel: "Med", NeedAmount: 100, AmountContributed: 0, Contributor: "Bob"},
			},
			want: domain.MonthlyFinancialReport{
				Year:  2025,
				Month: time.September,
				Needs: []domain.NeedReportBlock{
					{
						Need:       "2025-09-20 Med",
						NeedAmount: 100,
						Contributors: []domain.ContributorAmount{
							{Name: "Alice", Amount: 0},
							{Name: "Bob", Amount: 0},
						},
						Total: 0,
					},
				},
				Contributors: []domain.ContributorAmount{
					{Name: "Alice", Amount: 0},
					{Name: "Bob", Amount: 0},
				},
				Total: 0,
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

			// Normalize order before comparison
			order := []string{"Alice", "Bob", "Charlie"}
			for i := range rep.Needs {
				rep.Needs[i].Contributors = sortContributors(rep.Needs[i].Contributors, order)
				tt.want.Needs[i].Contributors = sortContributors(tt.want.Needs[i].Contributors, order)
			}
			rep.Contributors = sortContributors(rep.Contributors, order)
			tt.want.Contributors = sortContributors(tt.want.Contributors, order)

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
