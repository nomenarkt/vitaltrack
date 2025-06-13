// Package airtable handles Airtable API interactions.
package airtable

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/nomenarkt/vitaltrack/backend/internal/domain"
)

// Client talks to the Airtable REST API.
type Client struct {
	baseURL string
}

// NewClient returns a Client configured from environment variables.
func NewClient() *Client {
	_ = godotenv.Load()

	// Validate required environment variables to avoid runtime errors
	if os.Getenv("AIRTABLE_BASE_ID") == "" ||
		os.Getenv("AIRTABLE_MEDICINES_TABLE") == "" ||
		os.Getenv("AIRTABLE_ENTRIES_TABLE") == "" ||
		os.Getenv("AIRTABLE_TOKEN") == "" {
		panic("missing Airtable configuration: ensure AIRTABLE_BASE_ID, AIRTABLE_MEDICINES_TABLE, AIRTABLE_ENTRIES_TABLE and AIRTABLE_TOKEN are set")
	}

	baseURL := os.Getenv("AIRTABLE_API_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.airtable.com"
	}

	return &Client{baseURL: baseURL}
}

type airtableRecord[T any] struct {
	ID     string `json:"id"`
	Fields T      `json:"fields"`
}

type airtableResponse[T any] struct {
	Records []airtableRecord[T] `json:"records"`
}

// FetchMedicines retrieves all medicines from Airtable.
func (c *Client) FetchMedicines() ([]domain.Medicine, error) {
	url := fmt.Sprintf("%s/v0/%s/%s",
		c.baseURL,
		os.Getenv("AIRTABLE_BASE_ID"),
		os.Getenv("AIRTABLE_MEDICINES_TABLE"))

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+os.Getenv("AIRTABLE_TOKEN"))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := res.Body.Close(); cerr != nil {
			log.Printf("airtable response close error: %v", cerr)
		}
	}()

	body, _ := io.ReadAll(res.Body)

	var errCheck map[string]interface{}
	if json.Unmarshal(body, &errCheck) == nil {
		if errVal, exists := errCheck["error"]; exists {
			return nil, fmt.Errorf("airtable error: %v", errVal)
		}
	}

	var response airtableResponse[domain.Medicine]
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	var meds []domain.Medicine
	for _, rec := range response.Records {
		m := rec.Fields
		m.ID = rec.ID
		meds = append(meds, m)
	}
	return meds, nil
}

// FetchStockEntries retrieves all stock entry records from Airtable.
func (c *Client) FetchStockEntries() ([]domain.StockEntry, error) {
	url := fmt.Sprintf("%s/v0/%s/%s",
		c.baseURL,
		os.Getenv("AIRTABLE_BASE_ID"),
		os.Getenv("AIRTABLE_ENTRIES_TABLE"))

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+os.Getenv("AIRTABLE_TOKEN"))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := res.Body.Close(); cerr != nil {
			log.Printf("airtable response close error: %v", cerr)
		}
	}()

	body, _ := io.ReadAll(res.Body)

	var errCheck map[string]interface{}
	if json.Unmarshal(body, &errCheck) == nil {
		if errVal, exists := errCheck["error"]; exists {
			return nil, fmt.Errorf("airtable error: %v", errVal)
		}
	}

	var response airtableResponse[domain.StockEntry]
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	var entries []domain.StockEntry
	for _, rec := range response.Records {
		e := rec.Fields
		e.ID = rec.ID
		entries = append(entries, e)
	}
	return entries, nil
}

// CreateStockEntry adds a new stock entry record in Airtable.
func (c *Client) CreateStockEntry(entry domain.StockEntry) error {
	url := fmt.Sprintf("%s/v0/%s/%s",
		c.baseURL,
		os.Getenv("AIRTABLE_BASE_ID"),
		os.Getenv("AIRTABLE_ENTRIES_TABLE"))

	payload := map[string]any{
		"fields": map[string]any{
			"medicine_id": []string{entry.MedicineID},
			"quantity":    entry.Quantity,
			"unit":        entry.Unit,
			"date":        entry.Date.Format("2006-01-02"),
		},
	}

	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("AIRTABLE_TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := res.Body.Close(); cerr != nil {
			log.Printf("airtable response close error: %v", cerr)
		}
	}()

	if res.StatusCode >= 300 {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("airtable error: %s", string(b))
	}

	return nil
}

// UpdateForecastDate records the latest forecast date for a medicine in Airtable.
func (c *Client) UpdateForecastDate(medicineID string, forecastDate, updatedAt time.Time) error {
	url := fmt.Sprintf("%s/v0/%s/%s/%s",
		c.baseURL,
		os.Getenv("AIRTABLE_BASE_ID"),
		os.Getenv("AIRTABLE_MEDICINES_TABLE"),
		medicineID)

	payload := map[string]any{
		"fields": map[string]any{
			"forecast_out_of_stock_date": forecastDate.Format("2006-01-02"), // ✅ always date-only
			"forecast_last_updated":      updatedAt.Format("2006-01-02"),
		},
	}

	body, _ := json.Marshal(payload)
	log.Printf("🧪 PATCH Airtable: recordID=%s body=%s", medicineID, string(body))

	req, _ := http.NewRequest("PATCH", url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("AIRTABLE_TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := res.Body.Close(); cerr != nil {
			log.Printf("airtable response close error: %v", cerr)
		}
	}()

	if res.StatusCode >= 300 {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("airtable error: %s", string(b))
	}

	return nil
}

// UpdateMedicineLastAlertedDate saves the last alert date for a medicine.
func (c *Client) UpdateMedicineLastAlertedDate(medicineID string, date time.Time) error {
	url := fmt.Sprintf("%s/v0/%s/%s/%s",
		c.baseURL,
		os.Getenv("AIRTABLE_BASE_ID"),
		os.Getenv("AIRTABLE_MEDICINES_TABLE"),
		medicineID)

	payload := map[string]any{
		"fields": map[string]any{
			"last_alerted_date": date.Format("2006-01-02"),
		},
	}

	body, _ := json.Marshal(payload)
	log.Printf("🧪 PATCH Airtable: recordID=%s body=%s", medicineID, string(body))

	req, _ := http.NewRequest("PATCH", url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("AIRTABLE_TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := res.Body.Close(); cerr != nil {
			log.Printf("airtable response close error: %v", cerr)
		}
	}()

	b, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		log.Printf("airtable update failed: status=%d body=%s", res.StatusCode, string(b))
		return fmt.Errorf("airtable error: %s", string(b))
	}

	log.Printf("🆗 updated last_alerted_date response=%s", string(b))
	return nil
}

type airtableFinancialFields struct {
	Date              domain.FlexibleDate `json:"Date"`
	NeedLabel         string              `json:"NeedLabel"`
	NeedAmount        float64             `json:"NeedAmount"`
	AmountContributed float64             `json:"AmountContributed"`
	MonthTag          string              `json:"MonthTag"`
	Contributor       string              `json:"Contributor"`
}

// FetchFinancialEntries retrieves all financial entries for the given month.
func (c *Client) FetchFinancialEntries(year int, month time.Month) ([]domain.FinancialEntry, error) {
	query := url.QueryEscape(fmt.Sprintf("MonthTag=\"%04d-%02d\"", year, month))
	url := fmt.Sprintf("%s/v0/%s/%s?filterByFormula=%s",
		c.baseURL,
		os.Getenv("AIRTABLE_BASE_ID"),
		os.Getenv("AIRTABLE_FINANCIAL_TABLE"),
		query)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+os.Getenv("AIRTABLE_TOKEN"))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	log.Printf("🧾 Raw Airtable response: %s", string(body))

	var errCheck map[string]interface{}
	if json.Unmarshal(body, &errCheck) == nil {
		if errVal, exists := errCheck["error"]; exists {
			return nil, fmt.Errorf("airtable error: %v", errVal)
		}
	}

	// ✅ Use intermediate field struct
	var response airtableResponse[airtableFinancialFields]
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	var entries []domain.FinancialEntry
	for _, rec := range response.Records {
		f := rec.Fields
		// ✅ Defensive filter for test stability
		if f.MonthTag != fmt.Sprintf("%04d-%02d", year, month) {
			continue
		}
		entries = append(entries, domain.FinancialEntry{
			ID:                rec.ID,
			Date:              f.Date,
			NeedLabel:         f.NeedLabel,
			NeedAmount:        f.NeedAmount,
			AmountContributed: f.AmountContributed,
			MonthTag:          f.MonthTag,
			Contributor:       f.Contributor,
		})
	}

	return entries, nil
}
