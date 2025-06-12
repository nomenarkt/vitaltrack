// Package telegram provides Telegram Bot API helpers.
package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"io"
	"net/url"
	"sort"
	"time"

	"github.com/joho/godotenv"

	"github.com/nomenarkt/vitaltrack/backend/internal/domain"
	"github.com/nomenarkt/vitaltrack/backend/internal/logic/stockcalc"
	"github.com/nomenarkt/vitaltrack/backend/internal/util"
)

type Client struct {
	Token   string
	ChatID  string
	baseURL string
}

func NewClient() *Client {
	_ = godotenv.Load()

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	baseURL := os.Getenv("TELEGRAM_API_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.telegram.org"
	}
	if token == "" || chatID == "" {
		log.Fatal("missing Telegram configuration: TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID must be set")
	}

	return &Client{
		Token:   token,
		ChatID:  chatID,
		baseURL: baseURL,
	}
}

func (c *Client) SendTelegramMessage(msg string) error {
	log.Printf("📨 Sending Telegram: %s", msg)

	escaped := util.EscapeMarkdown(msg)

	payload := map[string]string{
		"chat_id":    c.ChatID,
		"text":       escaped,
		"parse_mode": "MarkdownV2",
	}

	body, _ := json.Marshal(payload)

	res, err := http.Post(
		c.baseURL+"/bot"+c.Token+"/sendMessage",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := res.Body.Close(); cerr != nil {
			log.Println("telegram response close error:", cerr)
		}
	}()

	if res.StatusCode >= 300 {
		return fmt.Errorf("telegram error status: %d", res.StatusCode)
	}

	return nil
}

type Update struct {
	UpdateID int `json:"update_id"`
	Message  struct {
		Text string `json:"text"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
	} `json:"message"`
}

type GetUpdatesResponse struct {
	OK     bool     `json:"ok"`
	Result []Update `json:"result"`
}

func (c *Client) PollForCommands(
	fetchData func() ([]domain.Medicine, []domain.StockEntry, error),
	reportFn func(year, month int) (domain.MonthlyFinancialReport, error),
) {
	var lastUpdateID int

	log.Println("📨 Telegram polling started...")
	for {
		time.Sleep(2 * time.Second)

		apiURL := fmt.Sprintf("%s/bot%s/getUpdates?timeout=10&offset=%d", c.baseURL, c.Token, lastUpdateID+1)
		resp, err := http.Get(apiURL)
		if err != nil {
			log.Println("Telegram polling error:", err)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		if err := resp.Body.Close(); err != nil {
			log.Println("Telegram response close error:", err)
		}

		var updates GetUpdatesResponse
		if err := json.Unmarshal(body, &updates); err != nil {
			log.Println("Failed to decode Telegram updates:", err)
			continue
		}
		if !updates.OK {
			log.Printf("Telegram API error status %d: %s", resp.StatusCode, string(body))
			continue
		}

		for _, update := range updates.Result {
			lastUpdateID = update.UpdateID
			switch {
			case update.Message.Text == "/stock":
				log.Println("🟡 /stock command triggered")
				go c.handleStockCommand(update.Message.Chat.ID, fetchData)
			case strings.HasPrefix(update.Message.Text, "/finance"):
				log.Println("🟡 /finance command triggered")
				year, month := time.Now().Year(), time.Now().Month()
				parts := strings.Fields(update.Message.Text)
				if len(parts) > 1 {
					if t, err := time.Parse("2006-01", parts[1]); err == nil {
						year, month = t.Year(), t.Month()
					}
				}
				go c.handleFinanceCommand(update.Message.Chat.ID, reportFn, year, month)
			}
		}
	}
}

func (c *Client) handleStockCommand(chatID int64, fetchData func() ([]domain.Medicine, []domain.StockEntry, error)) {
	meds, entries, err := fetchData()
	if err != nil {
		if err := c.sendTo(chatID, "\u26a0\ufe0f Failed to fetch stock data."); err != nil {
			log.Println("failed to send /stock response:", err)
		}
		return
	}

	log.Printf("📦 meds: %d, entries: %d", len(meds), len(entries))
	if len(meds) == 0 {
		if err := c.sendTo(chatID, "\u26a0\ufe0f No medicine or stock data found."); err != nil {
			log.Println("failed to send /stock response:", err)
		}
		return
	}

	now := time.Now()
	type Row struct {
		Name  string
		Date  time.Time
		Pills float64
	}
	var rows []Row
	for _, m := range meds {
		stock := stockcalc.CurrentStockAt(m, entries, now)
		if m.DailyDose == 0 || stock <= 0 {
			continue
		}
		date := stockcalc.OutOfStockDateAt(m, stock, now)
		rows = append(rows, Row{m.Name, date, stock})
	}

	if len(rows) == 0 {
		if err := c.sendTo(chatID, "\u2705 All medicines are well stocked."); err != nil {
			log.Println("failed to send /stock response:", err)
		}
		return
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Date.Before(rows[j].Date)
	})

	var lines []string
	for _, r := range rows {
		lines = append(lines, fmt.Sprintf("%-22s → %s (%.2f left)", r.Name, r.Date.Format("2006-01-02"), r.Pills))
	}

	msg := "*Out-of-Stock Forecast*\n\n```text\n" + strings.Join(lines, "\n") + "\n```"
	if err := c.sendTo(chatID, msg); err != nil {
		log.Println("failed to send /stock response:", err)
	} else {
		log.Println("sent /stock forecast")
	}
}

func (c *Client) handleFinanceCommand(chatID int64, fn func(year, month int) (domain.MonthlyFinancialReport, error), year int, month time.Month) {
	log.Printf("💸 Generating financial report for %d-%02d", year, month)
	report, err := fn(year, int(month))
	if err != nil {
		if err := c.sendTo(chatID, "\u26a0\ufe0f Failed to fetch financial data."); err != nil {
			log.Println("failed to send /finance response:", err)
		}
		return
	}

	var sections []string
	totalNeed := 0.0
	for _, n := range report.Needs {
		sections = append(sections, renderNeedBlock(n))
		totalNeed += n.NeedAmount
	}

	var summary []string
	summary = append(summary, "🧮 Monthly Summary")
	summary = append(summary, fmt.Sprintf("💰 Total Needs: %s", formatMGA(totalNeed)))
	summary = append(summary, fmt.Sprintf("💵 Total Contributed: %s", formatMGA(report.Total)))
	summary = append(summary, "")
	summary = append(summary, "👤 By Contributor:")
	for _, ctb := range report.Contributors {
		summary = append(summary, fmt.Sprintf("- %s \u2192 %s", ctb.Name, formatMGA(ctb.Amount)))
	}

	msg := fmt.Sprintf("*Financial Report %d-%02d*\n\n%s\n\n%s",
		report.Year, report.Month, strings.Join(sections, "\n\n"), strings.Join(summary, "\n"))

	if err := c.sendTo(chatID, msg); err != nil {
		log.Println("failed to send /finance response:", err)
	}
}

func (c *Client) sendTo(chatID int64, msg string) error {
	if msg == "" {
		return fmt.Errorf("empty telegram message")
	}

	escaped := util.EscapeMarkdown(msg)
	if len(escaped) > 4000 {
		escaped = escaped[:4000]
	}

	payload := url.Values{}
	payload.Set("chat_id", fmt.Sprintf("%d", chatID))
	payload.Set("text", escaped)
	payload.Set("parse_mode", "MarkdownV2")

	res, err := http.PostForm(
		c.baseURL+"/bot"+c.Token+"/sendMessage",
		payload,
	)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := res.Body.Close(); cerr != nil {
			log.Println("telegram response close error:", cerr)
		}
	}()

	if res.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(res.Body)
		log.Printf("telegram send failed: status=%d body=%s", res.StatusCode, string(body))
		return fmt.Errorf("telegram error status: %d", res.StatusCode)
	}

	return nil
}

// formatMGA formats numbers with comma separators and adds NARROW NO-BREAK SPACE before MGA.
func formatMGA(v float64) string {
	s := fmt.Sprintf("%.0f", v)
	n := len(s)
	if n <= 3 {
		return s + "\u202fMGA"
	}

	var out []rune
	for i, c := range s {
		if (n-i)%3 == 0 && i != 0 {
			out = append(out, ',')
		}
		out = append(out, c)
	}
	return string(out) + "\u202fMGA"
}

// renderNeedBlock formats a single need report block in monospaced layout.
func renderNeedBlock(n domain.NeedReportBlock) string {
	parts := strings.SplitN(n.Need, " ", 2)
	dateStr := parts[0]
	label := ""
	if len(parts) > 1 {
		label = parts[1]
	}
	d, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		d = time.Time{}
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("📅 %s – %s", d.Format("2006-01-02"), label))
	lines = append(lines, fmt.Sprintf("Need:          %s", formatMGA(n.NeedAmount)))
	lines = append(lines, fmt.Sprintf("Contributed:   %s", formatMGA(n.Total)))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("| %-12s | %-12s |", "Contributor", "Amount"))
	lines = append(lines, fmt.Sprintf("|%s|%s|", strings.Repeat("-", 14), strings.Repeat("-", 14)))

	// ❌ DO NOT re-sort here. Keep usecase-defined order.
	for _, ctb := range n.Contributors {
		lines = append(lines, fmt.Sprintf("| %-12s | %12s |", ctb.Name, formatMGA(ctb.Amount)))
	}

	return "```text\n" + strings.Join(lines, "\n") + "\n```"
}
