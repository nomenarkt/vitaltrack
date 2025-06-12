// Package domain contains core business models.
package domain

type Medicine struct {
	ID                     string        `json:"id"`
	Name                   string        `json:"name"`
	UnitType               string        `json:"unit_type"`
	UnitPerBox             float64       `json:"unit_per_box"`
	DailyDose              float64       `json:"daily_dose"`
	StartDate              FlexibleDate  `json:"start_date"` // supports RFC3339 + YYYY-MM-DD
	InitialStock           float64       `json:"initial_stock"`
	ForecastOutOfStockDate *FlexibleDate `json:"forecast_out_of_stock_date,omitempty"`
	ForecastLastUpdated    *FlexibleDate `json:"forecast_last_updated,omitempty"`
	LastAlertedDate        *FlexibleDate `json:"last_alerted_date,omitempty"`
}

type StockEntry struct {
	ID         string       `json:"id"`
	MedicineID string       `json:"medicine_id"`
	Quantity   float64      `json:"quantity"`
	Unit       string       `json:"unit"` // "box" or "pill"
	Date       FlexibleDate `json:"date"`
}
