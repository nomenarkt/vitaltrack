// File: internal/domain/models.go
package domain

import "time"

type Medicine struct {
	ID                     string     `json:"id"`
	Name                   string     `json:"name"`
	UnitType               string     `json:"unit_type"`
	UnitPerBox             float64    `json:"unit_per_box"`
	DailyDose              float64    `json:"daily_dose"`
	StartDate              string     `json:"start_date"` // keep as string for now
	InitialStock           float64    `json:"initial_stock"`
	ForecastOutOfStockDate *time.Time `json:"forecast_out_of_stock_date,omitempty"`
	ForecastLastUpdated    *time.Time `json:"forecast_last_updated,omitempty"`
}

type StockEntry struct {
	ID         string    `json:"id"`
	MedicineID string    `json:"medicine_id"`
	Quantity   int       `json:"quantity"`
	Unit       string    `json:"unit"` // "box" or "pill"
	Date       time.Time `json:"date"` // ISO8601: "2025-06-02"
}
