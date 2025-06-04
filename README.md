# 💊 Medicine Tracker

A reliable, real-time medicine stock tracker with:

- ✅ Dynamic pill stock forecasting
- 📆 Out-of-stock date prediction
- 🔁 Refill handling (idempotent & date-bound)
- 🤖 Telegram alerts
- 📊 Airtable-based data storage

---

## 🛠️ Stack

| Layer      | Tech                        |
|------------|-----------------------------|
| Backend    | Go (Fiber)                  |
| Data Store | Airtable (REST API)         |
| Alerts     | Telegram Bot API            |
| Deployment | GitHub Actions (soon)       |

---

## 🔧 Configuration

Create a `.env` file in `backend/` (based on `.env.template`):

```env
AIRTABLE_BASE_ID=
AIRTABLE_MEDICINES_TABLE=
AIRTABLE_ENTRIES_TABLE=
AIRTABLE_TOKEN=
TELEGRAM_BOT_TOKEN=
TELEGRAM_CHAT_ID=
ENABLE_ENTRY_POST=true
