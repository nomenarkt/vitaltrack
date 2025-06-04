package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type Client struct {
	Token  string
	ChatID string
}

func NewClient() *Client {
	_ = godotenv.Load()
	return &Client{
		Token:  os.Getenv("TELEGRAM_TOKEN"),
		ChatID: os.Getenv("TELEGRAM_CHAT_ID"),
	}
}

func (c *Client) SendTelegramMessage(msg string) error {
	payload := map[string]string{
		"chat_id":    c.ChatID,
		"text":       msg,
		"parse_mode": "Markdown",
	}

	body, _ := json.Marshal(payload)

	res, err := http.Post(
		"https://api.telegram.org/bot"+c.Token+"/sendMessage",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		return fmt.Errorf("telegram error status: %d", res.StatusCode)
	}

	return nil
}
