package background

import (
	"log"
	"time"
)

func StartStockAlertTicker(handler func() error) {
	go func() {
		for {
			log.Println("⏰ Checking stock for Telegram alerts...")
			if err := handler(); err != nil {
				log.Println("❌ Stock alert check failed:", err)
			} else {
				log.Println("✅ Stock alert check completed.")
			}
			time.Sleep(24 * time.Hour) //test: 1 * time.Minute
		}
	}()
}
