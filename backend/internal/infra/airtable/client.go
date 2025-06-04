package airtable

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/nomenarkt/medicine-tracker/backend/internal/domain"
)

type airtableRecord[T any] struct {
	ID     string `json:"id"`
	Fields T      `json:"fields"`
}

type airtableResponse[T any] struct {
	Records []airtableRecord[T] `json:"records"`
}

func loadEnv() {
	_ = godotenv.Load()
}

func FetchMedicines() ([]domain.Medicine, error) {
	loadEnv()
	fmt.Println("BASE ID =", os.Getenv("AIRTABLE_BASE_ID"))
	fmt.Println("TABLE =", os.Getenv("AIRTABLE_MEDICINES_TABLE"))

	url := fmt.Sprintf("https://api.airtable.com/v0/%s/%s", os.Getenv("AIRTABLE_BASE_ID"), os.Getenv("AIRTABLE_MEDICINES_TABLE"))
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+os.Getenv("AIRTABLE_TOKEN"))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	fmt.Println("=== RAW AIRTABLE JSON ===")
	fmt.Println(string(body))

	// Early check for Airtable error structure
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

	fmt.Println("Fetched medicine IDs:")
	for _, rec := range response.Records {
		fmt.Println("Airtable ID:", rec.ID, "| Field ID:", rec.Fields.ID)
	}

	meds := []domain.Medicine{}
	for _, rec := range response.Records {
		m := rec.Fields
		m.ID = rec.ID
		meds = append(meds, m)
	}
	return meds, nil
}

func FetchStockEntries() ([]domain.StockEntry, error) {
	loadEnv()
	url := fmt.Sprintf("https://api.airtable.com/v0/%s/%s", os.Getenv("AIRTABLE_BASE_ID"), os.Getenv("AIRTABLE_ENTRIES_TABLE"))

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+os.Getenv("AIRTABLE_TOKEN"))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	// Airtable error check
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

	entries := []domain.StockEntry{}
	for _, rec := range response.Records {
		e := rec.Fields
		e.ID = rec.ID
		entries = append(entries, e)
	}
	return entries, nil
}

//Internal adapter, useful for future automation (e.g. Telegram command /addstock)

func CreateStockEntry(entry domain.StockEntry) error {
	url := fmt.Sprintf("https://api.airtable.com/v0/%s/%s", os.Getenv("AIRTABLE_BASE_ID"), os.Getenv("AIRTABLE_ENTRIES_TABLE"))

	payload := map[string]any{
		"fields": map[string]any{
			"medicine_id": []string{entry.MedicineID}, // ✅ MUST be an array
			"quantity":    entry.Quantity,
			"unit":        entry.Unit,
			"date":        entry.Date,
		},
	}

	body, _ := json.Marshal(payload)
	fmt.Println("➡ Sending payload to Airtable:")
	fmt.Println(string(body)) // ✅ Log the outgoing JSON

	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("AIRTABLE_TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		b, _ := io.ReadAll(res.Body)
		fmt.Println("⛔ Airtable error response:")
		fmt.Println(string(b)) // ✅ Log the Airtable error response
		return fmt.Errorf("airtable error: %s", string(b))
	}

	fmt.Println("✅ Airtable record created successfully.")
	return nil
}

func UpdateForecastDate(medicineID string, forecastDate, updatedAt time.Time) error {
	loadEnv()

	url := fmt.Sprintf("https://api.airtable.com/v0/%s/%s/%s",
		os.Getenv("AIRTABLE_BASE_ID"),
		os.Getenv("AIRTABLE_MEDICINES_TABLE"),
		medicineID,
	)

	payload := map[string]any{
		"fields": map[string]any{
			"forecast_out_of_stock_date": forecastDate.Format("2006-01-02"),
			"forecast_last_updated":      updatedAt.Format("2006-01-02"),
		},
	}

	body, _ := json.Marshal(payload)
	fmt.Println("➡ Updating forecast in Airtable:")
	fmt.Println(string(body))

	req, _ := http.NewRequest("PATCH", url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("AIRTABLE_TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		b, _ := io.ReadAll(res.Body)
		fmt.Println("⛔ Airtable error response:")
		fmt.Println(string(b))
		return fmt.Errorf("airtable error: %s", string(b))
	}

	fmt.Println("✅ Forecast updated in Airtable for medicine:", medicineID)
	return nil
}
