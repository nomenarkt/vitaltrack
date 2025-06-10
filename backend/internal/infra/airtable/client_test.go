package airtable

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestUpdateMedicineLastAlertedDate(t *testing.T) {
	date := time.Date(2025, 6, 10, 0, 0, 0, 0, time.UTC)
	baseID := "app1"
	table := "Meds"
	recID := "rec1"

	var path string
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"id":"%s","fields":{"last_alerted_date":"%s"}}`, recID, date.Format("2006-01-02"))
	}))
	defer srv.Close()

	os.Setenv("AIRTABLE_BASE_ID", baseID)
	os.Setenv("AIRTABLE_MEDICINES_TABLE", table)
	os.Setenv("AIRTABLE_TOKEN", "tok")

	c := &Client{baseURL: srv.URL}

	var buf bytes.Buffer
	orig := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(orig)

	if err := c.UpdateMedicineLastAlertedDate(recID, date); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expPath := fmt.Sprintf("/v0/%s/%s/%s", baseID, table, recID)
	if path != expPath {
		t.Errorf("path = %s, want %s", path, expPath)
	}
	if !bytes.Contains(body, []byte(date.Format("2006-01-02"))) {
		t.Errorf("request body missing date: %s", string(body))
	}
	logs := buf.String()
	if !strings.Contains(logs, "PATCH Airtable") {
		t.Errorf("expected debug log, got %s", logs)
	}
	if !strings.Contains(logs, "response=") {
		t.Errorf("expected log with response, got %s", logs)
	}
}

func TestUpdateMedicineLastAlertedDate_ignoresMismatch(t *testing.T) {
	date := time.Date(2025, 6, 10, 0, 0, 0, 0, time.UTC)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"id":"rec","fields":{"last_alerted_date":"2024-01-01"}}`)
	}))
	defer srv.Close()

	os.Setenv("AIRTABLE_BASE_ID", "base")
	os.Setenv("AIRTABLE_MEDICINES_TABLE", "table")
	os.Setenv("AIRTABLE_TOKEN", "tok")

	c := &Client{baseURL: srv.URL}

	var buf bytes.Buffer
	orig := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(orig)

	if err := c.UpdateMedicineLastAlertedDate("rec", date); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "response=") {
		t.Errorf("expected log with response, got %s", buf.String())
	}
}

func TestUpdateMedicineLastAlertedDate_errorStatus(t *testing.T) {
	date := time.Date(2025, 6, 10, 0, 0, 0, 0, time.UTC)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "bad")
	}))
	defer srv.Close()

	os.Setenv("AIRTABLE_BASE_ID", "bid")
	os.Setenv("AIRTABLE_MEDICINES_TABLE", "tab")
	os.Setenv("AIRTABLE_TOKEN", "tok")

	c := &Client{baseURL: srv.URL}

	var buf bytes.Buffer
	orig := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(orig)

	err := c.UpdateMedicineLastAlertedDate("rec", date)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(buf.String(), "update failed") {
		t.Errorf("expected failure log, got %s", buf.String())
	}
}

func TestUpdateForecastDate(t *testing.T) {
	forecast := time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC)
	updated := time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC)
	baseID := "bid"
	table := "meds"
	recID := "rec9"

	var path string
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		body, _ = io.ReadAll(r.Body)
		fmt.Fprint(w, `{}`)
	}))
	defer srv.Close()

	os.Setenv("AIRTABLE_BASE_ID", baseID)
	os.Setenv("AIRTABLE_MEDICINES_TABLE", table)
	os.Setenv("AIRTABLE_TOKEN", "tok")

	c := &Client{baseURL: srv.URL}

	var buf bytes.Buffer
	orig := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(orig)

	if err := c.UpdateForecastDate(recID, forecast, updated); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expPath := fmt.Sprintf("/v0/%s/%s/%s", baseID, table, recID)
	if path != expPath {
		t.Errorf("path = %s, want %s", path, expPath)
	}
	logs := buf.String()
	if !strings.Contains(logs, "PATCH Airtable") {
		t.Errorf("expected debug log, got %s", logs)
	}
	if !bytes.Contains(body, []byte(forecast.Format("2006-01-02"))) {
		t.Errorf("request body missing forecast date: %s", string(body))
	}
	if !bytes.Contains(body, []byte(updated.Format("2006-01-02"))) {
		t.Errorf("request body missing updated date: %s", string(body))
	}
}

func TestFetchMedicines_AssignsID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"records":[{"id":"recA","fields":{"name":"MedA"}}]}`)
	}))
	defer srv.Close()

	os.Setenv("AIRTABLE_BASE_ID", "base")
	os.Setenv("AIRTABLE_MEDICINES_TABLE", "table")
	os.Setenv("AIRTABLE_TOKEN", "tok")

	c := &Client{baseURL: srv.URL}
	meds, err := c.FetchMedicines()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(meds) != 1 || meds[0].ID != "recA" {
		t.Fatalf("expected record ID set, got %+v", meds)
	}
}
