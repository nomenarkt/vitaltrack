# Medicine Tracker
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A Go-based Telegram-integrated medicine stock tracker to forecast depletion dates, notify on refill needs, and manage medicine inventory efficiently.

---

## 🚀 Features

* Forecast out-of-stock dates based on daily dosage and stock.
* `/stock` command to view medicine forecasts in Telegram.
* Automatic background alerts for upcoming refills (if enabled).
* Airtable backend integration.
* MarkdownV2-safe Telegram messages.

---

## 🛠️ Stack

| Layer      | Tech                      |
| ---------- | ------------------------- |
| Backend    | Go (Fiber)                |
| Data Store | Airtable (REST API)       |
| Alerts     | Telegram Bot API          |
| Deployment | Render (free-tier Docker) |

---

## 🧱 Why This Stack?

Our stack is optimized for simplicity, cost-efficiency, and real-time alerting.

| Layer          | Tech              | Why?                                                                                         |
| -------------- | ----------------- | -------------------------------------------------------------------------------------------- |
| **Backend**    | Go (Fiber)        | Fast, minimal memory footprint, perfect for APIs. Fiber is expressive but fast like Node.js. |
| **Data Store** | Airtable REST API | Great for prototyping with a spreadsheet-like UI and flexible schema.                        |
| **Alerts**     | Telegram Bot API  | Easy setup, excellent UX for push notifications, and no cost.                                |
| **Deploy**     | Render.com        | Free-tier hosting with Docker support, zero-config, and smooth scaling.                      |

This stack enables lean infrastructure with real-time automation and friendly UX — ideal for side projects, MVPs, and internal tools.

---

## Airtable Integration

Our Airtable base uses three tables:

### Medicines

* `id`
* `name`
* `unit_type`
* `unit_per_box`
* `daily_dose`
* `start_date`
* `initial_stock`
* `forecast_last_updated`
* `forecast_out_of_stock_date`
* `last_alerted_date`
* link to Stock Entries

### Stock Entries

* `id`
* `date`
* `quantity`
* `unit`
* `medicine_id`

### Financial Contributions

Tracks donations for each medicine.

The Airtable API token is supplied via the `AIRTABLE_TOKEN` environment variable. Never commit secrets to version control.

---

## ⚙️ Environment Variables

```
TELEGRAM_BOT_TOKEN=<your_token>
TELEGRAM_CHAT_ID=<target_chat_id>
TELEGRAM_API_BASE_URL=https://api.telegram.org

AIRTABLE_API_KEY=<airtable_key>
AIRTABLE_BASE_ID=<airtable_base>
AIRTABLE_MEDICINES_TABLE=Medicines
AIRTABLE_ENTRIES_TABLE=Entries
AIRTABLE_FINANCIAL_TABLE=<FinancialContributions table name>

ENABLE_ALERT_TICKER=true
ALERT_TICKER_INTERVAL=24h
ENABLE_TELEGRAM_POLLING=true
```

---

## 💬 Telegram Commands

### `/stock`

Generates a forecast for all medicines:

````
*Out-of-Stock Forecast*

```text
MedA                  → 2025-06-19 (20.00 left)
MedB                  → 2025-06-22 (6.50 left)
...etc
````

```

### Automatic Alerts (Ticker)
When enabled, the app sends alerts like:

```

⚠️ *Refill Alert* for *MedA* – runs out on *2025-06-19*
(20.00 pills left)

````

---

## 🧪 Testing

```bash
make test
````

Test coverage includes:

* ✅ Markdown escaping (`EscapeMarkdown`) for Telegram-safe messages.
* ✅ Forecast computation (initial stock, entry logs, fractional doses).
* ✅ Alert triggering based on threshold and schedule.
* ✅ `/stock` output formatting.
* ✅ Telegram delivery mock testing via `httptest`.

---

## 🏗️ Structure

```
backend/
├── internal/
│   ├── background/        # Ticker alert loop
│   ├── delivery/          # HTTP, CLI, Telegram handlers
│   ├── domain/            # Entities, logic types
│   ├── infra/             # Airtable + Telegram client
│   ├── usecase/           # Domain use cases
│   ├── util/              # Escape, formatting helpers
│   └── di/                # Dependency injection
├── Dockerfile
└── README.md
```

---

## 📦 Deployment

Supports Docker:

```bash
docker build -t medicine-tracker .
docker run -p 8787:8787 --env-file .env medicine-tracker
```

---

## 📣 Notes

* Ensure `ENABLE_ALERT_TICKER=true` and `ALERT_TICKER_INTERVAL` are set to trigger background alerts.
* Only **one instance** of polling should run to avoid Telegram `409 conflict` errors.
* Escape logic follows [MarkdownV2 rules](https://core.telegram.org/bots/api#markdownv2-style).

---

## 👥 Contributors

Thanks to everyone who contributed:

* [@nomenarkt](https://github.com/nomenarkt) – Creator & maintainer
* The Architect – Software Engineering Master GPT
* Codex by OpenAI

Your name here? Open a PR 😄

---

## 👨‍⚕️ Made for caregivers, by engineers.

Keep your loved ones’ medicine under control. ❤️
