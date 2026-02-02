package app

import "time"

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	AdminToken  string

	LogLevel string
	LogFile  string

	AnsweringSeconds time.Duration
	ResultsPause     time.Duration
	MaxRounds        int
}
