package background_test

import (
	"bytes"
	"log"
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

type mockTelegram struct{ msgs []string }

func (m *mockTelegram) SendTelegramMessage(msg string) error {
	m.msgs = append(m.msgs, msg)
	return nil
}
func (m *mockTelegram) PollForCommands(func() ([]domain.Medicine, []domain.StockEntry, error)) {}

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

			background.StartStockAlertTicker(deps, 10*time.Millisecond)
			time.Sleep(20 * time.Millisecond)

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
