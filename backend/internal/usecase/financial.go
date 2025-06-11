package usecase

import (
	"fmt"
	"sort"
	"time"

	"github.com/nomenarkt/medicine-tracker/backend/internal/domain"
	"github.com/nomenarkt/medicine-tracker/backend/internal/domain/ports"
)

// FinancialReportService handles aggregation of financial entries.
type FinancialReportService struct {
	Repo ports.FinancialDataPort
}

// GenerateFinancialReport groups financial entries by need and contributor.
func (s FinancialReportService) GenerateFinancialReport(year, month int) (domain.MonthlyFinancialReport, error) {
	entries, err := s.Repo.FetchFinancialEntries(year, time.Month(month))
	if err != nil {
		return domain.MonthlyFinancialReport{}, fmt.Errorf("fetch financial entries failed: %w", err)
	}

	needTotals := map[string]float64{}
	contributorTotals := map[string]float64{}
	total := 0.0

	for _, e := range entries {
		needTotals[e.Need] += e.Amount
		contributorTotals[e.Contributor] += e.Amount
		total += e.Amount
	}

	var needs []domain.NeedReport
	for n, v := range needTotals {
		needs = append(needs, domain.NeedReport{Need: n, Total: v})
	}
	sort.Slice(needs, func(i, j int) bool { return needs[i].Need < needs[j].Need })

	var contributors []domain.ContributorSummary
	for c, v := range contributorTotals {
		contributors = append(contributors, domain.ContributorSummary{Contributor: c, Total: v})
	}
	sort.Slice(contributors, func(i, j int) bool { return contributors[i].Contributor < contributors[j].Contributor })

	return domain.MonthlyFinancialReport{
		Year:         year,
		Month:        time.Month(month),
		Needs:        needs,
		Contributors: contributors,
		Total:        total,
	}, nil
}
