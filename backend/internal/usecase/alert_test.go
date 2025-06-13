package usecase_test

import (
	"bytes"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/nomenarkt/vitaltrack/backend/internal/domain"
	"github.com/nomenarkt/vitaltrack/backend/internal/usecase"
)

type mockAirtable struct {
	meds        []domain.Medicine
	entries     []domain.StockEntry
	updatedID   string
	updatedDate time.Time
}

func (m *mockAirtable) FetchMedicines() ([]domain.Medicine, error) {
	return m.meds, nil
}
func (m *mockAirtable) FetchStockEntries() ([]domain.StockEntry, error) {
	return m.entries, nil
}
func (m *mockAirtable) UpdateMedicineLastAlertedDate(medicineID string, date time.Time) error {
	m.updatedID = medicineID
	m.updatedDate = date
	return nil
}
func (m *mockAirtable) FetchFinancialEntries(int, time.Month) ([]domain.FinancialEntry, error) {
	return nil, nil
}

type mockTelegram struct {
	sent []string
}

func (m *mockTelegram) SendTelegramMessage(msg string) error {
	m.sent = append(m.sent, msg)
	return nil
}

func (m *mockTelegram) PollForCommands(_ func() ([]domain.Medicine, []domain.StockEntry, error), _ func(int, int) (domain.MonthlyFinancialReport, error)) {
	// no-op
}

func TestCheckAndAlertLowStock_Table(t *testing.T) {
	now := time.Now().UTC().Truncate(24 * time.Hour)

	tests := []struct {
		name        string
		med         domain.Medicine
		entries     []domain.StockEntry
		expectAlert bool
		expectText  string
	}{
		{
			name: "days11",
			med: domain.Medicine{
				ID:           "med11",
				Name:         "Med11",
				StartDate:    domain.NewFlexibleDate(now),
				InitialStock: 22,
				DailyDose:    2,
				UnitPerBox:   10,
			},
			expectAlert: false,
		},
		{
			name: "days10",
			med: domain.Medicine{
				ID:           "med10",
				Name:         "Med10",
				StartDate:    domain.NewFlexibleDate(now),
				InitialStock: 20,
				DailyDose:    2,
				UnitPerBox:   10,
			},
			expectAlert: true,
			expectText:  "*Med10* will run out",
		},
		{
			name: "days1",
			med: domain.Medicine{
				ID:           "med1",
				Name:         "Med1",
				StartDate:    domain.NewFlexibleDate(now),
				InitialStock: 2,
				DailyDose:    2,
				UnitPerBox:   10,
			},
			expectAlert: true,
			expectText:  "*Med1* will run out",
		},
		{
			name: "refill_today",
			med: domain.Medicine{
				ID:           "medr",
				Name:         "RefillMed",
				StartDate:    domain.NewFlexibleDate(now),
				InitialStock: 0,
				DailyDose:    2,
				UnitPerBox:   10,
			},
			entries: []domain.StockEntry{
				{
					MedicineID: "medr",
					Quantity:   2,
					Unit:       "box",
					Date:       domain.NewFlexibleDate(now),
				},
			},
			expectAlert: true,
			expectText:  "*Refill recorded for RefillMed*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			at := &mockAirtable{
				meds:    []domain.Medicine{tt.med},
				entries: tt.entries,
			}
			tg := &mockTelegram{}

			checker := usecase.StockChecker{
				Airtable: at,
				Telegram: tg,
			}

			err := checker.CheckAndAlertLowStock()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expectAlert {
				found := false
				for _, msg := range tg.sent {
					if strings.Contains(msg, tt.expectText) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected alert for %s, but none sent", tt.med.Name)
				}
			} else {
				for _, msg := range tg.sent {
					if strings.Contains(msg, tt.med.Name) {
						t.Errorf("unexpected alert for %s", tt.med.Name)
					}
				}
			}
		})
	}
}

func TestCheckAndAlertLowStock_UpdatesLastAlerted(t *testing.T) {
	now := time.Now().UTC().Truncate(24 * time.Hour)

	at := &mockAirtable{
		meds: []domain.Medicine{
			{
				ID:           "rec99",
				Name:         "Med99",
				StartDate:    domain.NewFlexibleDate(now),
				InitialStock: 10,
				DailyDose:    1,
				UnitPerBox:   10,
			},
		},
	}
	tg := &mockTelegram{}
	checker := usecase.StockChecker{Airtable: at, Telegram: tg}

	var buf bytes.Buffer
	orig := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(orig)

	if err := checker.CheckAndAlertLowStock(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if at.updatedID != "rec99" {
		t.Errorf("expected update for rec99, got %s", at.updatedID)
	}
	logs := buf.String()
	if !strings.Contains(logs, "Calling UpdateMedicineLastAlertedDate") {
		t.Errorf("missing update log: %s", logs)
	}
}
