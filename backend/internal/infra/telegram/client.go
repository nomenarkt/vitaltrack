// File: internal/infra/telegram/client.go
package telegram

import (
	"bytes"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func loadEnv() {
	_ = godotenv.Load()
}

func SendTelegramMessage(msg string) error {
	loadEnv()

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	if token == "" || chatID == "" {
		return fmt.Errorf("missing TELEGRAM_BOT_TOKEN or TELEGRAM_CHAT_ID")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	payload := []byte(fmt.Sprintf(`{"chat_id": "%s", "text": "%s"}`, chatID, msg))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		return fmt.Errorf("telegram API error: status %d", res.StatusCode)
	}

	return nil
}
