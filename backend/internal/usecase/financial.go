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

	breakdown := map[string]map[string]float64{}
	contributorTotals := map[string]float64{}
	contributorSet := map[string]struct{}{}
	total := 0.0

	for _, e := range entries {
		key := fmt.Sprintf("%s %s", e.Date.Format("2006-01-02"), e.NeedLabel)
		if _, ok := breakdown[key]; !ok {
			breakdown[key] = map[string]float64{}
		}
		breakdown[key][e.Contributor] += e.Amount
		contributorTotals[e.Contributor] += e.Amount
		contributorSet[e.Contributor] = struct{}{}
		total += e.Amount
	}

	var contributorNames []string
	for c := range contributorSet {
		contributorNames = append(contributorNames, c)
	}
	sort.Strings(contributorNames)

	var needKeys []string
	for k := range breakdown {
		needKeys = append(needKeys, k)
	}
	sort.Strings(needKeys)

	var needs []domain.NeedReportBlock
	for _, k := range needKeys {
		contribMap := breakdown[k]
		var contribs []domain.ContributorAmount
		var needTotal float64
		for _, name := range contributorNames {
			amt := contribMap[name]
			contribs = append(contribs, domain.ContributorAmount{Name: name, Amount: amt})
			needTotal += amt
		}
		needs = append(needs, domain.NeedReportBlock{Need: k, Contributors: contribs, Total: needTotal})
	}

	var contributors []domain.ContributorAmount
	for _, name := range contributorNames {
		contributors = append(contributors, domain.ContributorAmount{Name: name, Amount: contributorTotals[name]})
	}

	return domain.MonthlyFinancialReport{
		Year:         year,
		Month:        time.Month(month),
		Needs:        needs,
		Contributors: contributors,
		Total:        total,
	}, nil
}
