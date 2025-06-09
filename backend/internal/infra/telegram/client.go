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

	"github.com/nomenarkt/medicine-tracker/backend/internal/domain"
	"github.com/nomenarkt/medicine-tracker/backend/internal/logic/stockcalc"
	"github.com/nomenarkt/medicine-tracker/backend/internal/util"
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

func (c *Client) PollForCommands(fetchData func() ([]domain.Medicine, []domain.StockEntry, error)) {
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
			if update.Message.Text == "/stock" {
				log.Println("🟡 /stock command triggered")
				go c.handleStockCommand(update.Message.Chat.ID, fetchData)
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
