package telegram

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nomenarkt/medicine-tracker/backend/internal/domain"
	"github.com/nomenarkt/medicine-tracker/backend/internal/usecase"
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
			expect:  "*Out-of-Stock Forecast*",
		},
		{
			name:    "all_good",
			meds:    []domain.Medicine{{ID: "m2", Name: "Med2", StartDate: domain.NewFlexibleDate(now), InitialStock: 0, DailyDose: 1, UnitPerBox: 10}},
			entries: []domain.StockEntry{{MedicineID: "m2", Quantity: 1.0, Unit: "box", Date: domain.NewFlexibleDate(now)}},
			expect:  "*Out-of-Stock Forecast*",
		},
		{
			name:    "forecast",
			meds:    []domain.Medicine{{ID: "m3", Name: "Med3", StartDate: domain.NewFlexibleDate(now), InitialStock: 10, DailyDose: 1, UnitPerBox: 10}},
			entries: []domain.StockEntry{{MedicineID: "m3", Quantity: 1.0, Unit: "box", Date: domain.NewFlexibleDate(now)}},
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
			if !strings.Contains(got, expected) {
				t.Errorf("expected %q in message, got %q", tt.expect, got)
			}

			if !strings.Contains(logBuf.String(), "📦 meds:") {
				t.Errorf("expected log of data counts")
			}
		})
	}
}

func TestHandleStockCommand_onlyInitialStock(t *testing.T) {
	srv, msgs := newTestServer()
	defer srv.Close()

	now := time.Now().AddDate(0, 0, -2)
	meds := []domain.Medicine{{ID: "m4", Name: "InitOnly", StartDate: domain.NewFlexibleDate(now), InitialStock: 5, DailyDose: 1, UnitPerBox: 10}}
	fetch := func() ([]domain.Medicine, []domain.StockEntry, error) {
		return meds, []domain.StockEntry{}, nil
	}

	c := &Client{Token: "test", ChatID: "1", baseURL: srv.URL}
	c.handleStockCommand(456, fetch)

	if len(*msgs) == 0 {
		t.Fatalf("no telegram message sent")
	}

	got := (*msgs)[0]
	expected := util.EscapeMarkdown("*Out-of-Stock Forecast*")
	if !strings.Contains(got, expected) {
		t.Errorf("expected forecast message, got %q", got)
	}
}

func TestHandleStockCommand_withFloatEntries(t *testing.T) {
	srv, msgs := newTestServer()
	defer srv.Close()

	now := time.Now().AddDate(0, 0, -1)
	meds := []domain.Medicine{{ID: "m5", Name: "FloatMed", StartDate: domain.NewFlexibleDate(now), InitialStock: 0, DailyDose: 1, UnitPerBox: 10}}
	entries := []domain.StockEntry{{MedicineID: "m5", Quantity: 0.75, Unit: "box", Date: domain.NewFlexibleDate(now)}}
	fetch := func() ([]domain.Medicine, []domain.StockEntry, error) {
		return meds, entries, nil
	}

	c := &Client{Token: "test", ChatID: "1", baseURL: srv.URL}
	c.handleStockCommand(789, fetch)

	if len(*msgs) == 0 {
		t.Fatalf("no telegram message sent")
	}

	got := (*msgs)[0]
	expected := util.EscapeMarkdown("*Out-of-Stock Forecast*")
	if !strings.Contains(got, expected) {
		t.Errorf("expected forecast message, got %q", got)
	}
}

func TestHandleStockCommand_zeroDose(t *testing.T) {
	srv, msgs := newTestServer()
	defer srv.Close()

	now := time.Now().AddDate(0, 0, -1)
	meds := []domain.Medicine{{ID: "m6", Name: "ZeroDose", StartDate: domain.NewFlexibleDate(now), InitialStock: 10, DailyDose: 0, UnitPerBox: 10}}
	entries := []domain.StockEntry{{MedicineID: "m6", Quantity: 1.0, Unit: "box", Date: domain.NewFlexibleDate(now)}}
	fetch := func() ([]domain.Medicine, []domain.StockEntry, error) {
		return meds, entries, nil
	}

	c := &Client{Token: "test", ChatID: "1", baseURL: srv.URL}
	c.handleStockCommand(999, fetch)

	if len(*msgs) == 0 {
		t.Fatalf("no telegram message sent")
	}

	got := (*msgs)[0]
	expected := util.EscapeMarkdown("\u2705 All medicines are well stocked.")
	if !strings.Contains(got, expected) {
		t.Errorf("expected all-stocked message, got %q", got)
	}
}

func TestHandleStockCommand_partialDose(t *testing.T) {
	srv, msgs := newTestServer()
	defer srv.Close()

	now := time.Now().UTC().Truncate(24 * time.Hour)
	tests := []struct {
		name         string
		daily        float64
		initialStock float64
	}{
		{name: "quarter", daily: 0.25, initialStock: 1},
		{name: "half", daily: 0.5, initialStock: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meds := []domain.Medicine{{ID: "mp" + tt.name, Name: "Partial" + tt.name, StartDate: domain.NewFlexibleDate(now), InitialStock: tt.initialStock, DailyDose: tt.daily, UnitPerBox: 1}}
			fetch := func() ([]domain.Medicine, []domain.StockEntry, error) {
				return meds, []domain.StockEntry{}, nil
			}

			c := &Client{Token: "test", ChatID: "1", baseURL: srv.URL}
			c.handleStockCommand(100, fetch)

			if len(*msgs) == 0 {
				t.Fatalf("no telegram message sent")
			}
			got := (*msgs)[len(*msgs)-1]
			expectedDate := util.EscapeMarkdown(now.AddDate(0, 0, int(math.Floor(tt.initialStock/tt.daily))).Format("2006-01-02"))
			if !strings.Contains(got, expectedDate) {
				t.Errorf("expected date %s in message, got %q", expectedDate, got)
			}
		})
	}
}

func TestHandleStockCommand_refillAppliedCumulatively(t *testing.T) {
	srv, msgs := newTestServer()
	defer srv.Close()

	now := time.Now().UTC().Truncate(24 * time.Hour)
	start := now.AddDate(0, 0, -5)
	entries := []domain.StockEntry{
		{MedicineID: "mref", Quantity: 1, Unit: "box", Date: domain.NewFlexibleDate(start.AddDate(0, 0, 1))},
		{MedicineID: "mref", Quantity: 1, Unit: "box", Date: domain.NewFlexibleDate(start.AddDate(0, 0, 2))},
		{MedicineID: "mref", Quantity: 1, Unit: "box", Date: domain.NewFlexibleDate(start.AddDate(0, 0, 3))},
	}
	meds := []domain.Medicine{{ID: "mref", Name: "Refill", StartDate: domain.NewFlexibleDate(start), InitialStock: 0, DailyDose: 1, UnitPerBox: 10}}
	fetch := func() ([]domain.Medicine, []domain.StockEntry, error) { return meds, entries, nil }

	c := &Client{Token: "test", ChatID: "1", baseURL: srv.URL}
	c.handleStockCommand(200, fetch)

	if len(*msgs) == 0 {
		t.Fatalf("no telegram message sent")
	}
	got := (*msgs)[0]
	stock := 30.0 - 5.0 // 3 boxes = 30, consumed 5
	days := int(math.Floor(stock))
	expectedDate := util.EscapeMarkdown(now.AddDate(0, 0, days).Format("2006-01-02"))
	if !strings.Contains(got, expectedDate) {
		t.Errorf("expected cumulative forecast date %s, got %q", expectedDate, got)
	}
}

func TestHandleStockCommand_fetchError(t *testing.T) {
	srv, msgs := newTestServer()
	defer srv.Close()

	fetch := func() ([]domain.Medicine, []domain.StockEntry, error) {
		return nil, nil, fmt.Errorf("boom")
	}
	c := &Client{Token: "test", ChatID: "1", baseURL: srv.URL}
	c.handleStockCommand(300, fetch)

	if len(*msgs) == 0 {
		t.Fatalf("no telegram message sent")
	}
	if (*msgs)[0] != util.EscapeMarkdown("\u26a0\ufe0f Failed to fetch stock data.") {
		t.Errorf("expected fetch error message, got %q", (*msgs)[0])
	}
}

type mockFinanceRepo struct{ entries []domain.FinancialEntry }

func (m mockFinanceRepo) FetchFinancialEntries(year int, month time.Month) ([]domain.FinancialEntry, error) {
	return m.entries, nil
}

func TestHandleFinanceCommand(t *testing.T) {
	srv, msgs := newTestServer()
	defer srv.Close()

	date := time.Date(2025, 6, 5, 0, 0, 0, 0, time.UTC)
	entries := []domain.FinancialEntry{
		{Date: domain.NewFlexibleDate(date), Need: "Med", Contributor: "Bob", Amount: 5},
		{Date: domain.NewFlexibleDate(date), Need: "Med", Contributor: "Alice", Amount: 10},
		{Date: domain.NewFlexibleDate(date), Need: "Med", Contributor: "Charlie", Amount: 0},
	}

	svc := usecase.FinancialReportService{Repo: mockFinanceRepo{entries: entries}}
	report, err := svc.GenerateFinancialReport(2025, int(time.June))
	if err != nil {
		t.Fatalf("generate report error: %v", err)
	}

	fn := func(y, m int) (domain.MonthlyFinancialReport, error) { return report, nil }
	c := &Client{Token: "tok", ChatID: "1", baseURL: srv.URL}
	c.handleFinanceCommand(55, fn, 2025, time.June)

	if len(*msgs) == 0 {
		t.Fatalf("no telegram message sent")
	}
	msg := (*msgs)[0]

	header := util.EscapeMarkdown("*Financial Report 2025-06*")
	if !strings.Contains(msg, header) {
		t.Fatalf("unexpected message: %s", msg)
	}

	emojis := []string{"📅", "💊", "💰", "🧮", "👤", "💵"}
	for _, e := range emojis {
		if !strings.Contains(msg, e) {
			t.Errorf("missing %s in message", e)
		}
	}

	expectedOrder := []string{
		util.EscapeMarkdown("- Alice \u2192 10\u202FMGA"),
		util.EscapeMarkdown("- Bob \u2192 5\u202FMGA"),
		util.EscapeMarkdown("- Charlie \u2192 0\u202FMGA"),
	}
	last := -1
	for _, e := range expectedOrder {
		idx := strings.Index(msg, e)
		if idx == -1 {
			t.Fatalf("missing %s in message", e)
		}
		if idx < last {
			t.Errorf("contributors not sorted alphabetically")
		}
		last = idx
	}

	row := util.EscapeMarkdown("| Charlie | 0\u202FMGA |")
	if !strings.Contains(msg, row) {
		t.Errorf("expected zero contribution row, got %s", msg)
	}

	if !strings.Contains(msg, "\u202fMGA") {
		t.Errorf("narrow space missing in MGA values")
	}
}

func TestSendTo_escapesMarkdown(t *testing.T) {
	srv, msgs := newTestServer()
	defer srv.Close()

	c := &Client{Token: "test", ChatID: "1", baseURL: srv.URL}

	name := "NEBI-LOL 5mg (sample)"
	if err := c.sendTo(111, name); err != nil {
		t.Fatalf("sendTo error: %v", err)
	}

	if len(*msgs) == 0 {
		t.Fatalf("no telegram message sent")
	}

	expected := util.EscapeMarkdown(name)
	if (*msgs)[0] != expected {
		t.Errorf("expected %q, got %q", expected, (*msgs)[0])
	}
}
