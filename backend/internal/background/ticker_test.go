package background_test

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/nomenarkt/medicine-tracker/backend/internal/background"
	"github.com/nomenarkt/medicine-tracker/backend/internal/di"
	"github.com/nomenarkt/medicine-tracker/backend/internal/domain"
)

type mockAirtable struct {
	meds    []domain.Medicine
	entries []domain.StockEntry
}

func (m *mockAirtable) FetchMedicines() ([]domain.Medicine, error)            { return m.meds, nil }
func (m *mockAirtable) FetchStockEntries() ([]domain.StockEntry, error)       { return m.entries, nil }
func (m *mockAirtable) CreateStockEntry(domain.StockEntry) error              { return nil }
func (m *mockAirtable) UpdateForecastDate(string, time.Time, time.Time) error { return nil }
func (m *mockAirtable) FetchFinancialEntries(int, time.Month) ([]domain.FinancialEntry, error) {
	return nil, nil
}

type mockTelegram struct{ msgs []string }

func (m *mockTelegram) SendTelegramMessage(msg string) error {
	m.msgs = append(m.msgs, msg)
	return nil
}
func (m *mockTelegram) PollForCommands(func() ([]domain.Medicine, []domain.StockEntry, error), func(int, int) (domain.MonthlyFinancialReport, error)) {
}

type httpTelegram struct {
	url    string
	posted *[]string
}

func (h *httpTelegram) SendTelegramMessage(msg string) error {
	_, err := http.PostForm(h.url, url.Values{"text": {msg}})
	if err == nil {
		*h.posted = append(*h.posted, msg)
	}
	return err
}

func (h *httpTelegram) PollForCommands(func() ([]domain.Medicine, []domain.StockEntry, error), func(int, int) (domain.MonthlyFinancialReport, error)) {
}

func TestStartStockAlertTicker(t *testing.T) {
	now := time.Date(2025, 6, 4, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		med    domain.Medicine
		expect bool
	}{
		{
			name:   "should_send",
			med:    domain.Medicine{ID: "m1", Name: "Med1", StartDate: domain.NewFlexibleDate(now.AddDate(0, 0, -3)), InitialStock: 20, DailyDose: 2, UnitPerBox: 10},
			expect: true,
		},
		{
			name:   "skip_high_stock",
			med:    domain.Medicine{ID: "m2", Name: "Med2", StartDate: domain.NewFlexibleDate(now.AddDate(0, 0, -1)), InitialStock: 50, DailyDose: 1, UnitPerBox: 10},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			at := &mockAirtable{meds: []domain.Medicine{tt.med}, entries: []domain.StockEntry{}}
			tg := &mockTelegram{}
			deps := di.Dependencies{Airtable: at, Telegram: tg}

			var buf bytes.Buffer
			orig := log.Writer()
			log.SetOutput(&buf)
			defer log.SetOutput(orig)

			stop := background.StartStockAlertTicker(deps, 10*time.Millisecond, func() time.Time { return now })
			time.Sleep(20 * time.Millisecond)
			stop()

			if tt.expect && len(tg.msgs) == 0 {
				t.Fatalf("expected alert but none sent")
			}
			if !tt.expect && len(tg.msgs) > 0 {
				t.Fatalf("unexpected alert sent: %v", tg.msgs)
			}

			if !strings.Contains(buf.String(), "🔁 Alert ticker completed") {
				t.Errorf("expected completion log")
			}
		})
	}
}

func TestStartStockAlertTicker_HTTP(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name   string
		med    domain.Medicine
		expect bool
	}{
		{
			name:   "alert",
			med:    domain.Medicine{ID: "a1", Name: "AlertMed", StartDate: domain.NewFlexibleDate(now.AddDate(0, 0, -15)), InitialStock: 20, DailyDose: 1, UnitPerBox: 10},
			expect: true,
		},
		{
			name:   "skip",
			med:    domain.Medicine{ID: "a2", Name: "SkipMed", StartDate: domain.NewFlexibleDate(now.AddDate(0, 0, -1)), InitialStock: 50, DailyDose: 1, UnitPerBox: 10},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			posted := []string{}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.ReadAll(r.Body)
				posted = append(posted, r.URL.Path)
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			at := &mockAirtable{meds: []domain.Medicine{tt.med}, entries: []domain.StockEntry{}}
			tg := &httpTelegram{url: srv.URL, posted: &posted}
			deps := di.Dependencies{Airtable: at, Telegram: tg}

			var buf bytes.Buffer
			orig := log.Writer()
			log.SetOutput(&buf)
			defer log.SetOutput(orig)

			stop := background.StartStockAlertTicker(deps, 10*time.Millisecond, func() time.Time { return now })
			time.Sleep(20 * time.Millisecond)
			stop()

			if tt.expect && len(posted) == 0 {
				t.Fatalf("expected alert POST")
			}
			if !tt.expect && len(posted) > 0 {
				t.Fatalf("unexpected alert sent: %v", posted)
			}
			if !strings.Contains(buf.String(), "Alert ticker completed") {
				t.Errorf("expected completion log")
			}
		})
	}
}
