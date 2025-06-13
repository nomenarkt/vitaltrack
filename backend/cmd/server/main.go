// Command server starts the HTTP server.
package main

import (
	"log"

	"github.com/joho/godotenv"

	_ "github.com/nomenarkt/vitaltrack/backend/internal/background"
	"github.com/nomenarkt/vitaltrack/backend/internal/di"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("godotenv load: %v", err)
	}

	app := di.NewApp()

	log.Fatal(app.Listen(":8787"))
}
