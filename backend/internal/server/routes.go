// Package server provides HTTP route bindings for all service endpoints.
package server

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/nomenarkt/vitaltrack/backend/internal/domain"
	"github.com/nomenarkt/vitaltrack/backend/internal/domain/ports"
	"github.com/nomenarkt/vitaltrack/backend/internal/usecase"
)

// SetupRoutes registers all HTTP endpoints with the provided Fiber app.
func SetupRoutes(
	app *fiber.App,
	checker *usecase.StockChecker,
	forecastSvc usecase.OutOfStockService,
	medicineSvc usecase.MedicineService,
	dataPort ports.StockDataPort,
	telegramClient ports.TelegramService,
) {
	const stockThreshold = 10.0
	allowEntryPost := os.Getenv("ENABLE_ENTRY_POST") == "true"

	// ✅ New route for manual stock check via HTTP
	app.Get("/check", func(c *fiber.Ctx) error {
		if err := checker.CheckAndAlertLowStock(); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "ok"})
	})

	app.Get("/debug/medicines", func(c *fiber.Ctx) error {
		meds, err := dataPort.FetchMedicines()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(meds)
	})

	app.Get("/debug/entries", func(c *fiber.Ctx) error {
		entries, err := dataPort.FetchStockEntries()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(entries)
	})

	app.Get("/api/medicines/:id/stock", func(c *fiber.Ctx) error {
		id := c.Params("id")
		now := time.Now().UTC()

		info, err := medicineSvc.GetStockInfo(id, now)
		if err != nil {
			if errors.Is(err, usecase.ErrMedicineNotFound) {
				return c.Status(404).JSON(fiber.Map{"error": err.Error()})
			}
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		if info.CurrentStock < stockThreshold {
			alert := fmt.Sprintf("⚠️ Stock alert for *%s*:\nOnly %.2f pills left!\nRefill before %s.",
				id,
				info.CurrentStock,
				info.OutOfStockDate.Format("Jan 2, 2006"),
			)
			if err := telegramClient.SendTelegramMessage(alert); err != nil {
				log.Printf("telegram send error: %v", err)
			}
		}

		return c.JSON(fiber.Map{
			"initial_stock":     info.InitialStock,
			"consumed_stock":    info.ConsumedStock,
			"current_stock":     info.CurrentStock,
			"out_of_stock_date": info.OutOfStockDate.Format("2006-01-02"),
		})
	})

	app.Get("/debug/outofstock", func(c *fiber.Ctx) error {
		msg, err := forecastSvc.GenerateOutOfStockForecastMessage()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		if err := telegramClient.SendTelegramMessage(msg); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"message": "out-of-stock forecast sent"})
	})

	if allowEntryPost {
		app.Post("/api/medicines/:id/entries", func(c *fiber.Ctx) error {
			id := c.Params("id")

			var req domain.CreateStockEntryRequest
			if err := c.BodyParser(&req); err != nil {
				return c.Status(400).JSON(fiber.Map{"error": "invalid JSON body"})
			}
			if req.Quantity <= 0 || (req.Unit != "box" && req.Unit != "pill") || req.Date == "" {
				return c.Status(400).JSON(fiber.Map{
					"error": "quantity must be > 0, unit must be 'box' or 'pill', date must not be empty",
				})
			}

			parsedDate, err := time.Parse("2006-01-02", req.Date)
			if err != nil {
				parsedDate, err = time.Parse(time.RFC3339, req.Date)
				if err != nil {
					return c.Status(400).JSON(fiber.Map{
						"error": "invalid date format, expected YYYY-MM-DD or RFC3339",
					})
				}
			}

			entry := domain.StockEntry{
				MedicineID: []string{id},
				Quantity:   req.Quantity,
				Unit:       req.Unit,
				Date:       domain.FlexibleDate{Time: parsedDate},
			}
			if err := dataPort.CreateStockEntry(entry); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": err.Error()})
			}
			return c.Status(201).JSON(fiber.Map{"message": "stock entry created"})
		})
	}
}
