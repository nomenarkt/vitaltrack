package domain

import "time"

// FinancialEntry represents a single contribution toward a need.
type FinancialEntry struct {
	ID                string       `json:"id"`
	Date              FlexibleDate `json:"Date"`
	NeedLabel         string       `json:"NeedLabel"`
	NeedAmount        float64      `json:"NeedAmount"`
	AmountContributed float64      `json:"AmountContributed"`
	MonthTag          string       `json:"MonthTag"`
	Contributor       string       `json:"Contributor"`
}

// ContributorAmount represents the amount contributed by a single contributor.
type ContributorAmount struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

// NeedReportBlock aggregates contributions for a given need.
type NeedReportBlock struct {
	Need         string              `json:"need"`
	NeedAmount   float64             `json:"need_amount"`
	Contributors []ContributorAmount `json:"contributors"`
	Total        float64             `json:"total"`
}

// MonthlyFinancialReport summarises all financial entries for a month.
type MonthlyFinancialReport struct {
	Year         int
	Month        time.Month
	Needs        []NeedReportBlock
	Contributors []ContributorAmount
	Total        float64
}
