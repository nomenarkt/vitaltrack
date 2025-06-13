package background_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"strings"

	"github.com/nomenarkt/vitaltrack/backend/internal/background"
	"github.com/nomenarkt/vitaltrack/backend/internal/di"
	"github.com/nomenarkt/vitaltrack/backend/internal/domain"
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
	resp, err := http.PostForm(h.url, url.Values{"text": {msg}})
	if err != nil {
		return err
	}
	if err := resp.Body.Close(); err != nil {
		return err
	}
	*h.posted = append(*h.posted, msg)
	return nil
}

func (h *httpTelegram) PollForCommands(func() ([]domain.Medicine, []domain.StockEntry, error), func(int, int) (domain.MonthlyFinancialReport, error)) {
}

type captureLogger struct{ entries []string }

func (c *captureLogger) Info(_ context.Context, msg string, kv ...any) {
	c.entries = append(c.entries, logFmt(msg, kv...))
}

func (c *captureLogger) Error(_ context.Context, msg string, kv ...any) {
	c.entries = append(c.entries, logFmt(msg, kv...))
}

func (c *captureLogger) String() string { return strings.Join(c.entries, "\n") }

func logFmt(msg string, kv ...any) string {
	if len(kv) == 0 {
		return msg
	}
	parts := make([]string, 0, len(kv)/2)
	for i := 0; i < len(kv); i += 2 {
		k := fmt.Sprint(kv[i])
		v := ""
		if i+1 < len(kv) {
			v = fmt.Sprint(kv[i+1])
		}
		parts = append(parts, k+"="+v)
	}
	return msg + " " + strings.Join(parts, " ")
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
			lg := &captureLogger{}
			deps := di.Dependencies{Airtable: at, Telegram: tg, Logger: lg}

			stop := background.StartStockAlertTicker(context.Background(), deps, 10*time.Millisecond, func() time.Time { return now })
			time.Sleep(20 * time.Millisecond)
			stop()

			if tt.expect && len(tg.msgs) == 0 {
				t.Fatalf("expected alert but none sent")
			}
			if !tt.expect && len(tg.msgs) > 0 {
				t.Fatalf("unexpected alert sent: %v", tg.msgs)
			}
			if !strings.Contains(lg.String(), "alert ticker completed") {
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
				if _, err := io.ReadAll(r.Body); err != nil {
					t.Errorf("read body: %v", err)
				}
				posted = append(posted, r.URL.Path)
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			at := &mockAirtable{meds: []domain.Medicine{tt.med}, entries: []domain.StockEntry{}}
			tg := &httpTelegram{url: srv.URL, posted: &posted}
			lg := &captureLogger{}
			deps := di.Dependencies{Airtable: at, Telegram: tg, Logger: lg}

			stop := background.StartStockAlertTicker(context.Background(), deps, 10*time.Millisecond, func() time.Time { return now })
			time.Sleep(20 * time.Millisecond)
			stop()

			if tt.expect && len(posted) == 0 {
				t.Fatalf("expected alert POST")
			}
			if !tt.expect && len(posted) > 0 {
				t.Fatalf("unexpected alert sent: %v", posted)
			}
			if !strings.Contains(lg.String(), "alert ticker completed") {
				t.Errorf("expected completion log")
			}
		})
	}
}
