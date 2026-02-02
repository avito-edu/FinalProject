package main

import (
	"os"
	"time"

	"github.com/ArtemMoroz51/FinalProject/internal/app"
)

func main() {
	cfg := app.Config{
		HTTPAddr:    getenv("HTTP_ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		AdminToken:  os.Getenv("ADMIN_TOKEN"),

		LogLevel: getenv("LOG_LEVEL", "info"),
		LogFile:  getenv("LOG_FILE", "/tmp/app.log"),

		AnsweringSeconds: 30 * time.Second,
		ResultsPause:     5 * time.Second,
		MaxRounds:        5,
	}

	if cfg.DatabaseURL == "" {
		panic("DATABASE_URL is required")
	}

	a, err := app.New(cfg)
	if err != nil {
		panic(err)
	}
	defer a.Close()

	if err := a.Run(); err != nil {
		panic(err)
	}
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}
