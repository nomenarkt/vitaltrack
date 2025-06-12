# VitalTrack
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

VitalTrack is a Go-powered, Telegram-integrated system that tracks essential medical stock levels and financial contributions. Built for caregivers, nonprofits, and teams coordinating shared health responsibilities, it forecasts depletion, alerts you ahead of time, and reports donation progress — all in one clean bot interface.

---

## 🚀 Features

- Forecast medicine depletion dates based on daily dosage and stock.
- `/stock` Telegram command to view real-time forecasts.
- `/finance` command to view contribution summaries by month.
- Automatic alert ticker for refills (optional).
- Markdown-safe output for Telegram's MarkdownV2 format.
- Airtable as a simple no-code backend.
- Fully tested and CI-integrated.

---

## 🛠️ Stack

| Layer      | Tech                      |
| ---------- | ------------------------- |
| Backend    | Go (Fiber)                |
| Data Store | Airtable (REST API)       |
| Alerts     | Telegram Bot API          |
| Deployment | Docker via Render         |

---

## 🧱 Why This Stack?

| Layer          | Tech              | Why?                                                                                         |
| -------------- | ----------------- | -------------------------------------------------------------------------------------------- |
| **Backend**    | Go (Fiber)        | High performance and small memory footprint. Fiber is expressive, fast, and reliable.        |
| **Data Store** | Airtable REST API | Great for MVPs with a spreadsheet UI and flexible schema for non-technical users.            |
| **Alerts**     | Telegram Bot API  | Push-based, reliable, and great UX.                                                         |
| **Deploy**     | Render or Docker  | Fast zero-config container deployment (free tier ready).                                     |

---

## ⚙️ Environment Variables

```env
TELEGRAM_BOT_TOKEN=<your_token>
TELEGRAM_CHAT_ID=<target_chat_id>
TELEGRAM_API_BASE_URL=https://api.telegram.org

AIRTABLE_API_KEY=<airtable_key>
AIRTABLE_BASE_ID=<airtable_base>
AIRTABLE_API_HOST=https://api.airtable.com/v0
AIRTABLE_MEDICINES_TABLE=Medicines
AIRTABLE_ENTRIES_TABLE=Entries
AIRTABLE_FINANCIAL_TABLE=FinancialContributions

ENABLE_ALERT_TICKER=true
ALERT_TICKER_INTERVAL=24h
ENABLE_TELEGRAM_POLLING=true

💬 Telegram Commands
/stock
Returns a forecast for all tracked medicines:
*Out-of-Stock Forecast*

```text
MedA                  → 2025-06-19 (20.00 left)
MedB                  → 2025-06-22 (6.50 left)

### `/finance`
Returns a monthly contribution summary, per medicine and contributor:

Financial Report 2025-06
📅 2025-06-05 – Med
Need:          20 MGA
Contributed:   15 MGA

| Contributor  | Amount       |
|--------------|--------------|
| Alice        |      10 MGA |
| Bob          |       5 MGA |
| Charlie      |       0 MGA |

🧮 Monthly Summary
💰 Total Needs: 20 MGA
💵 Total Contributed: 15 MGA
👤 Contributor Summary:
-Alice → 10 MGA
-Bob → 5 MGA


---

## 🧪 Testing

```bash
make test

Test coverage includes:
✅ Markdown escaping for Telegram safety
✅ Forecast logic (stock/dose/day)
✅ Alert ticker
✅ /stock and /finance Telegram responses
✅ Financial aggregation by month and contributor

📦 Deployment
docker build -t vitaltrack .
docker run -p 8787:8787 --env-file .env vitaltrack

Or use Render with Dockerfile auto-deploy.

🧱 Codebase Structure
├── vitaltrack/
│   ├── vitaltrack/
│   │   ├── LICENSE
│   │   ├── README.md
│   │   ├── backend-contributing.md
│   │   ├── frontend-contributing.md
│   │   ├── vitaltrack.code-workspace
│   │   ├── .gitignore
│   │   ├── Dockerfile
│   │   ├── docker-compose.yml
│   │   ├── .dockerignore
│   │   ├── backend/
│   │   │   ├── .env.template
│   │   │   ├── .env
│   │   │   ├── go.mod
│   │   │   ├── go.sum
│   │   │   ├── Makefile
│   │   │   ├── internal/
│   │   │   │   ├── domain/
│   │   │   │   │   ├── flexible_date.go
│   │   │   │   │   ├── entry.go
│   │   │   │   │   ├── models.go
│   │   │   │   │   ├── financial.go
│   │   │   │   │   ├── ports/
│   │   │   │   │   │   ├── services.go
│   │   │   │   ├── usecase/
│   │   │   │   │   ├── alert.go
│   │   │   │   │   ├── alert_test.go
│   │   │   │   │   ├── financial.go
│   │   │   │   │   ├── financial_test.go
│   │   │   │   ├── infra/
│   │   │   │   │   ├── airtable/
│   │   │   │   │   │   ├── client.go
│   │   │   │   │   │   ├── client_test.go
│   │   │   │   │   ├── telegram/
│   │   │   │   │   │   ├── client.go
│   │   │   │   │   │   ├── client_test.go
│   │   │   │   ├── logic/
│   │   │   │   │   ├── forecast/
│   │   │   │   │   │   ├── forecast.go
│   │   │   │   │   │   ├── forecast_test.go
│   │   │   │   │   ├── stockcalc/
│   │   │   │   │   │   ├── stock.go
│   │   │   │   │   │   ├── stock_test.go
│   │   │   │   ├── util/
│   │   │   │   │   ├── escape.go
│   │   │   │   │   ├── escape_test.go
│   │   │   │   ├── delivery/
│   │   │   │   │   ├── telegram/
│   │   │   │   │   │   ├── handler.go
│   │   │   │   │   │   ├── handler_test.go
│   │   │   │   ├── background/
│   │   │   │   │   ├── ticker.go
│   │   │   │   │   ├── ticker_test.go
│   │   │   │   ├── di/
│   │   │   │   │   ├── wire.go
│   │   │   │   │   ├── wire_gen.go
│   │   │   │   ├── server/
│   │   │   │   │   ├── server.go
│   │   │   │   │   ├── main.go

🔐 CI/CD & Linting
CI runs on GitHub Actions:
-✅ Unit tests with Go 1.24.1
-✅ Build binary
-✅ golangci-lint v1.64.8 checks
-✅ Docker build
-🧪 Coverage tracking (via future Codecov or upload logic)
Trigger: [push, pull_request]
See .github/workflows/ci.yml.

👥 Contributors
Thanks to everyone who contributed:
-@nomenarkt — Creator & Maintainer
-The Architect — Software Engineering Master GPT
-Codex — Implementation engine by OpenAI

❤️ Built for communities
Helping you track what's vital — be it health, time, or donations.