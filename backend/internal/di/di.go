package di

import (
	"github.com/nomenarkt/medicine-tracker/backend/internal/domain/ports"
	"github.com/nomenarkt/medicine-tracker/backend/internal/infra/airtable"
	"github.com/nomenarkt/medicine-tracker/backend/internal/infra/telegram"
	"github.com/nomenarkt/medicine-tracker/backend/internal/usecase"
)

type Dependencies struct {
	Airtable     ports.StockDataPort // satisfies AirtableService + StockDataPort
	Telegram     ports.TelegramService
	StockChecker *usecase.StockChecker
	ForecastSvc  usecase.OutOfStockService
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
	}
}
