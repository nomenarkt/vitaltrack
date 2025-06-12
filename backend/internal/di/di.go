package di

import (
	"github.com/nomenarkt/vitaltrack/backend/internal/domain/ports"
	"github.com/nomenarkt/vitaltrack/backend/internal/infra/airtable"
	"github.com/nomenarkt/vitaltrack/backend/internal/infra/telegram"
	"github.com/nomenarkt/vitaltrack/backend/internal/usecase"
)

type Dependencies struct {
	Airtable     ports.StockDataPort // satisfies AirtableService + StockDataPort
	Telegram     ports.TelegramService
	StockChecker *usecase.StockChecker
	ForecastSvc  usecase.OutOfStockService
	FinancialSvc usecase.FinancialReportService
}

func Init() Dependencies {
	at := airtable.NewClient()
	tg := telegram.NewClient()

	return Dependencies{
		Airtable: at,
		Telegram: tg,
		StockChecker: &usecase.StockChecker{
			Airtable: at,
			Telegram: tg,
		},
		ForecastSvc: usecase.OutOfStockService{
			Airtable: at,
		},
		FinancialSvc: usecase.FinancialReportService{Repo: at},
	}
}
