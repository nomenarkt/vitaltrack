package stockcalc_test

import (
	"testing"
	"time"

	"github.com/nomenarkt/vitaltrack/backend/internal/domain"
	"github.com/nomenarkt/vitaltrack/backend/internal/logic/stockcalc"
)

func mustDate(s string) domain.FlexibleDate {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return domain.NewFlexibleDate(t)
}

func TestCurrentStockAt_WithRefillOnToday(t *testing.T) {
	now := time.Date(2025, 6, 4, 0, 0, 0, 0, time.UTC)

	med := domain.Medicine{
		ID:           "med123",
		Name:         "Paracetamol",
		StartDate:    domain.NewFlexibleDate(time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)),
		InitialStock: 10,
		DailyDose:    1,
		UnitPerBox:   10,
	}

	entries := []domain.StockEntry{
		{
			MedicineID: []string{"med123"},
			Quantity:   1.0,
			Unit:       "box",
			Date:       domain.NewFlexibleDate(now),
		},
	}

	got := stockcalc.CurrentStockAt(med, entries, now)
	want := float64(10 - 3 + 10) // used 3 doses (June 2,3,4) + 1 box refill

	if got != want {
		t.Errorf("Expected stock %.2f, got %.2f", want, got)
	}
}

func TestCurrentStockAt_WithMultipleEntryDates(t *testing.T) {
	med := domain.Medicine{
		ID:           "med1",
		Name:         "TestMed",
		UnitPerBox:   10,
		DailyDose:    1.0,
		StartDate:    domain.NewFlexibleDate(time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)),
		InitialStock: 5,
	}

	today, err := time.Parse("2006-01-02", "2025-06-04")
	if err != nil {
		t.Fatalf("parse date: %v", err)
	}

	entries := []domain.StockEntry{
		{MedicineID: []string{"med1"}, Quantity: 1.0, Unit: "box", Date: domain.NewFlexibleDate(today)},                    // +10
		{MedicineID: []string{"med1"}, Quantity: 5.0, Unit: "pill", Date: domain.NewFlexibleDate(today)},                   // +5
		{MedicineID: []string{"med1"}, Quantity: 5.0, Unit: "pill", Date: domain.NewFlexibleDate(today.AddDate(0, 0, 1))},  // future: ignored
		{MedicineID: []string{"med1"}, Quantity: 5.0, Unit: "pill", Date: domain.NewFlexibleDate(today.AddDate(0, 0, -1))}, // past: included
	}

	stock := stockcalc.CurrentStockAt(med, entries, today)

	expected := 5.0 + 20.0 - 3.0 // initial + all past refills - consumed
	if stock != expected {
		t.Errorf("Expected %.2f, got %.2f", expected, stock)
	}
}

func TestOutOfStockDateAt(t *testing.T) {
	now := time.Date(2025, 6, 4, 0, 0, 0, 0, time.UTC)

	med := domain.Medicine{
		ID:        "med123",
		DailyDose: 2,
	}

	stock := 10.0
	got := stockcalc.OutOfStockDateAt(med, stock, now)
	want := now.AddDate(0, 0, 5)

	if !got.Equal(want) {
		t.Errorf("Expected out-of-stock date %v, got %v", want, got)
	}
}

func TestCurrentStockAt_WithRFC3339StartDate(t *testing.T) {
	now := time.Date(2025, 6, 4, 0, 0, 0, 0, time.UTC)

	med := domain.Medicine{
		ID:           "medRFC",
		Name:         "RFCMed",
		StartDate:    domain.NewFlexibleDate(time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)),
		InitialStock: 10,
		DailyDose:    1,
		UnitPerBox:   10,
	}

	entries := []domain.StockEntry{
		{
			MedicineID: []string{"medRFC"},
			Quantity:   1.0,
			Unit:       "box",
			Date:       domain.NewFlexibleDate(now),
		},
	}

	got := stockcalc.CurrentStockAt(med, entries, now)
	want := float64(10 - 3 + 10) // used 3 doses + 1 box refill

	if got != want {
		t.Errorf("Expected stock %.2f, got %.2f", want, got)
	}
}

func TestCurrentStockAt_EntryDateRFC3339Match(t *testing.T) {
	start := "2025-06-01"
	now := time.Date(2025, 6, 4, 12, 0, 0, 0, time.UTC)

	med := domain.Medicine{
		ID:           "med2",
		Name:         "AdvancedMed",
		StartDate:    mustDate(start),
		InitialStock: 5,
		DailyDose:    1,
		UnitPerBox:   10,
	}

	rfcDate := time.Date(2025, 6, 4, 12, 0, 0, 0, time.UTC)

	entries := []domain.StockEntry{
		{MedicineID: []string{"med2"}, Quantity: 1.0, Unit: "box", Date: domain.NewFlexibleDate(rfcDate)},
	}

	got := stockcalc.CurrentStockAt(med, entries, now)
	want := 5.0 - 3.0 + 10.0

	if got != want {
		t.Errorf("Expected %.2f, got %.2f", want, got)
	}
}
