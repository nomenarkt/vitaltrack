package domain

type CreateStockEntryRequest struct {
	Quantity int    `json:"quantity"`
	Unit     string `json:"unit"` // "pill" or "box"
	Date     string `json:"date"` // "2025-06-02"
}
