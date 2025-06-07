package telegram

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nomenarkt/medicine-tracker/backend/internal/domain"
	"github.com/nomenarkt/medicine-tracker/backend/internal/util"
)

func newTestServer() (*httptest.Server, *[]string) {
	msgs := &[]string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		*msgs = append(*msgs, r.Form.Get("text"))
		w.WriteHeader(http.StatusOK)
	}))
	return server, msgs
}

func TestHandleStockCommand(t *testing.T) {
	now := time.Now().AddDate(0, 0, -5)

	tests := []struct {
		name    string
		meds    []domain.Medicine
		entries []domain.StockEntry
		expect  string
	}{
		{
			name:    "no_data",
			meds:    []domain.Medicine{},
			entries: []domain.StockEntry{},
			expect:  "\u26a0\ufe0f No medicine or stock data found.",
		},
		{
			name:    "no_entries",
			meds:    []domain.Medicine{{ID: "m1", Name: "Med1", StartDate: domain.NewFlexibleDate(now), InitialStock: 10, DailyDose: 1, UnitPerBox: 10}},
			entries: []domain.StockEntry{},
			expect:  "\u26a0\ufe0f No medicine or stock data found.",
		},
		{
			name:    "all_good",
			meds:    []domain.Medicine{{ID: "m2", Name: "Med2", StartDate: domain.NewFlexibleDate(now), InitialStock: 0, DailyDose: 1, UnitPerBox: 10}},
			entries: []domain.StockEntry{{MedicineID: "m2", Quantity: 1, Unit: "box", Date: domain.NewFlexibleDate(now)}},
			expect:  "\u2705 All medicines are well stocked.",
		},
		{
			name:    "forecast",
			meds:    []domain.Medicine{{ID: "m3", Name: "Med3", StartDate: domain.NewFlexibleDate(now), InitialStock: 10, DailyDose: 1, UnitPerBox: 10}},
			entries: []domain.StockEntry{{MedicineID: "m3", Quantity: 1, Unit: "box", Date: domain.NewFlexibleDate(now)}},
			expect:  "*Out-of-Stock Forecast*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, msgs := newTestServer()
			defer srv.Close()

			var logBuf bytes.Buffer
			orig := log.Writer()
			log.SetOutput(&logBuf)
			defer log.SetOutput(orig)

			c := &Client{Token: "test", ChatID: "1", baseURL: srv.URL}
			fetch := func() ([]domain.Medicine, []domain.StockEntry, error) {
				return tt.meds, tt.entries, nil
			}
			c.handleStockCommand(123, fetch)

			if len(*msgs) == 0 {
				t.Fatalf("no telegram message sent")
			}
			got := (*msgs)[0]
			expected := util.EscapeMarkdown(tt.expect)
			if strings.Contains(tt.name, "forecast") {
				expected = tt.expect
			}
			if !strings.Contains(got, expected) {
				t.Errorf("expected %q in message, got %q", tt.expect, got)
			}

			if !strings.Contains(logBuf.String(), "/stock data") {
				t.Errorf("expected log of data counts")
			}
		})
	}
}
