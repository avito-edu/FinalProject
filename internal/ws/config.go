package ws

import "time"

const (
	answeringSeconds = 30 * time.Second
	resultsPause     = 5 * time.Second
	maxRounds        = 5

	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 8 * 1024
)
