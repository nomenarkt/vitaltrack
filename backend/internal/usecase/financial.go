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

	needBreakdown := map[string]map[string]float64{}
	contributorTotals := map[string]float64{}
	total := 0.0

	for _, e := range entries {
		if _, ok := needBreakdown[e.Need]; !ok {
			needBreakdown[e.Need] = map[string]float64{}
		}
		needBreakdown[e.Need][e.Contributor] += e.Amount
		contributorTotals[e.Contributor] += e.Amount
		total += e.Amount
	}

	var needs []domain.NeedReportBlock
	var needNames []string
	for n := range needBreakdown {
		needNames = append(needNames, n)
	}
	sort.Strings(needNames)
	for _, n := range needNames {
		contributorsMap := needBreakdown[n]
		var contribs []domain.ContributorAmount
		var needTotal float64
		for c, v := range contributorsMap {
			contribs = append(contribs, domain.ContributorAmount{Name: c, Amount: v})
			needTotal += v
		}
		sort.Slice(contribs, func(i, j int) bool { return contribs[i].Name < contribs[j].Name })
		needs = append(needs, domain.NeedReportBlock{Need: n, Contributors: contribs, Total: needTotal})
	}

	var contributors []domain.ContributorAmount
	for c, v := range contributorTotals {
		contributors = append(contributors, domain.ContributorAmount{Name: c, Amount: v})
	}
	sort.Slice(contributors, func(i, j int) bool { return contributors[i].Name < contributors[j].Name })

	return domain.MonthlyFinancialReport{
		Year:         year,
		Month:        time.Month(month),
		Needs:        needs,
		Contributors: contributors,
		Total:        total,
	}, nil
}
