package main

import (
	"log"

	"github.com/joho/godotenv"

	_ "github.com/nomenarkt/vitaltrack/backend/internal/background"
	"github.com/nomenarkt/vitaltrack/backend/internal/di"
)

func main() {
	_ = godotenv.Load()

	app := di.NewApp()

	log.Fatal(app.Listen(":8787"))
}
