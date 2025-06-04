package stockcalc_test

import (
	"testing"
	"time"

	"github.com/nomenarkt/medicine-tracker/backend/internal/domain"
	"github.com/nomenarkt/medicine-tracker/backend/internal/logic/stockcalc"
)

func TestCurrentStockAt_WithRefills(t *testing.T) {
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
	med := domain.Medicine{
		DailyDose: 2.0,
	}

	today := time.Date(2025, 6, 4, 0, 0, 0, 0, time.UTC)
	date := stockcalc.OutOfStockDateAt(med, 6.0, today)

	expected := today.AddDate(0, 0, 3)
	if !date.Equal(expected) {
		t.Errorf("Expected %s, got %s", expected.Format("2006-01-02"), date.Format("2006-01-02"))
	}
}
