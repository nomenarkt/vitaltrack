// Package di wires application dependencies for runtime use.
package di

import (
	"github.com/nomenarkt/vitaltrack/backend/internal/domain/ports"
	"github.com/nomenarkt/vitaltrack/backend/internal/infra/airtable"
	"github.com/nomenarkt/vitaltrack/backend/internal/infra/telegram"
	"github.com/nomenarkt/vitaltrack/backend/internal/logger"
	"github.com/nomenarkt/vitaltrack/backend/internal/usecase"
)

type Dependencies struct {
	Airtable     ports.StockDataPort // satisfies AirtableService + StockDataPort
	Telegram     ports.TelegramService
	Logger       logger.Logger
	StockChecker *usecase.StockChecker
	ForecastSvc  usecase.OutOfStockService
	FinancialSvc usecase.FinancialReportService
	MedicineSvc  usecase.MedicineService
}

func Init() Dependencies {
	at := airtable.NewClient()
	tg := telegram.NewClient()
	lg := logger.NewStdLogger()

	return Dependencies{
		Airtable: at,
		Telegram: tg,
		Logger:   lg,
		StockChecker: &usecase.StockChecker{
			Airtable: at,
			Telegram: tg,
		},
		ForecastSvc: usecase.OutOfStockService{
			Airtable: at,
		},
		FinancialSvc: usecase.FinancialReportService{Repo: at},
		MedicineSvc:  usecase.MedicineService{Repo: at},
	}
}
