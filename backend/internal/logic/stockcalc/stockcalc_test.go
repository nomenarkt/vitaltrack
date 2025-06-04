package stockcalc_test

import (
	"testing"
	"time"

	"github.com/nomenarkt/medicine-tracker/backend/internal/domain"
	"github.com/nomenarkt/medicine-tracker/backend/internal/logic/stockcalc"
)

func TestCurrentStockAt_WithRefillOnToday(t *testing.T) {
	now := time.Date(2025, 6, 4, 0, 0, 0, 0, time.UTC)

	med := domain.Medicine{
		ID:           "med123",
		Name:         "Paracetamol",
		StartDate:    "2025-06-01",
		InitialStock: 10,
		DailyDose:    1,
		UnitPerBox:   10,
	}

	entries := []domain.StockEntry{
		{
			MedicineID: "med123",
			Quantity:   1,
			Unit:       "box",
			Date:       now,
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
		StartDate:    "2025-06-01",
		InitialStock: 5,
	}

	today, _ := time.Parse("2006-01-02", "2025-06-04")

	entries := []domain.StockEntry{
		{MedicineID: "med1", Quantity: 1, Unit: "box", Date: today},                    // +10
		{MedicineID: "med1", Quantity: 5, Unit: "pill", Date: today},                   // +5
		{MedicineID: "med1", Quantity: 5, Unit: "pill", Date: today.AddDate(0, 0, 1)},  // future: ignored
		{MedicineID: "med1", Quantity: 5, Unit: "pill", Date: today.AddDate(0, 0, -1)}, // past: ignored
	}

	stock := stockcalc.CurrentStockAt(med, entries, today)

	expected := 5.0 - 3.0 + 15.0 // initial - 3 days used + today's refill
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
