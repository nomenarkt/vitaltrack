package domain

import "time"

// FinancialEntry represents a single contribution toward a need.
type FinancialEntry struct {
	ID          string       `json:"id"`
	Date        FlexibleDate `json:"date"`
	Need        string       `json:"need"`
	Contributor string       `json:"contributor"`
	Amount      float64      `json:"amount"`
}

// NeedReport aggregates total contributions per need.
type NeedReport struct {
	Need  string
	Total float64
}

// ContributorSummary aggregates total contributions per contributor.
type ContributorSummary struct {
	Contributor string
	Total       float64
}

// MonthlyFinancialReport summarises all financial entries for a month.
type MonthlyFinancialReport struct {
	Year         int
	Month        time.Month
	Needs        []NeedReport
	Contributors []ContributorSummary
	Total        float64
}
