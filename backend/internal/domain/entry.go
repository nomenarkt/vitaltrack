package domain

// CreateStockEntryRequest defines the payload for creating a stock entry.
type CreateStockEntryRequest struct {
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"` // "pill" or "box"
	Date     string  `json:"date"` // "2025-06-02"
}
