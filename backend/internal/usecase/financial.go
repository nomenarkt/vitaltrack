package usecase

import (
	"fmt"
	"sort"
	"time"

	"github.com/nomenarkt/vitaltrack/backend/internal/domain"
	"github.com/nomenarkt/vitaltrack/backend/internal/domain/ports"
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

	// 👥 Define contributor display order
	desiredOrder := []string{"Onja", "Tafita", "Henintsoa", "Mahandry"}
	orderIndex := map[string]int{}
	for i, name := range desiredOrder {
		orderIndex[name] = i
	}

	breakdown := map[string]map[string]float64{}
	needAmounts := map[string]float64{}
	needSeen := map[string]bool{}
	contributorTotals := map[string]float64{}
	contributorPresent := map[string]bool{}
	total := 0.0

	for _, e := range entries {
		key := fmt.Sprintf("%s %s", e.Date.Format("2006-01-02"), e.NeedLabel)

		if _, ok := breakdown[key]; !ok {
			breakdown[key] = map[string]float64{}
		}
		breakdown[key][e.Contributor] += e.AmountContributed

		if !needSeen[key] {
			needAmounts[key] = e.NeedAmount
			needSeen[key] = true
		}

		contributorTotals[e.Contributor] += e.AmountContributed
		contributorPresent[e.Contributor] = true
		total += e.AmountContributed
	}

	var contributorNames []string
	for _, name := range desiredOrder {
		if contributorPresent[name] {
			contributorNames = append(contributorNames, name)
		}
	}
	for name := range contributorPresent {
		if _, exists := orderIndex[name]; !exists {
			orderIndex[name] = len(orderIndex)
			contributorNames = append(contributorNames, name)
		}
	}

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
		// ⬇️ Enforce consistent contributor sort
		sort.SliceStable(contribs, func(i, j int) bool {
			return orderIndex[contribs[i].Name] < orderIndex[contribs[j].Name]
		})
		needs = append(needs, domain.NeedReportBlock{
			Need:         k,
			NeedAmount:   needAmounts[k],
			Contributors: contribs,
			Total:        needTotal,
		})
	}

	var contributors []domain.ContributorAmount
	for _, name := range contributorNames {
		contributors = append(contributors, domain.ContributorAmount{Name: name, Amount: contributorTotals[name]})
	}
	sort.SliceStable(contributors, func(i, j int) bool {
		return orderIndex[contributors[i].Name] < orderIndex[contributors[j].Name]
	})

	return domain.MonthlyFinancialReport{
		Year:         year,
		Month:        time.Month(month),
		Needs:        needs,
		Contributors: contributors,
		Total:        total,
	}, nil
}
