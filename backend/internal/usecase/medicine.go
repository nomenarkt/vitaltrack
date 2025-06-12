package usecase

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/nomenarkt/vitaltrack/backend/internal/domain"
	"github.com/nomenarkt/vitaltrack/backend/internal/domain/ports"
	"github.com/nomenarkt/vitaltrack/backend/internal/logic/stockcalc"
)

// MedicineService provides stock related operations.
type MedicineService struct {
	Repo ports.StockDataPort
}

var ErrMedicineNotFound = errors.New("medicine not found")

// StockInfo summarizes current stock information for a medicine.
type StockInfo struct {
	InitialStock   float64
	ConsumedStock  float64
	CurrentStock   float64
	OutOfStockDate time.Time
}

// GetStockInfo computes current stock and forecast for the given medicine.
func (s MedicineService) GetStockInfo(id string, now time.Time) (StockInfo, error) {
	meds, err := s.Repo.FetchMedicines()
	if err != nil {
		return StockInfo{}, fmt.Errorf("fetch medicines failed: %w", err)
	}
	entries, err := s.Repo.FetchStockEntries()
	if err != nil {
		return StockInfo{}, fmt.Errorf("fetch stock entries failed: %w", err)
	}

	var med *domain.Medicine
	for _, m := range meds {
		if m.ID == id {
			tmp := m
			med = &tmp
			break
		}
	}
	if med == nil {
		return StockInfo{}, ErrMedicineNotFound
	}

	stock := stockcalc.CurrentStockAt(*med, entries, now)
	forecast := stockcalc.OutOfStockDateAt(*med, stock, now)

	info := StockInfo{
		InitialStock:   med.InitialStock,
		ConsumedStock:  math.Max(med.InitialStock-stock, 0),
		CurrentStock:   stock,
		OutOfStockDate: forecast,
	}
	return info, nil
}
