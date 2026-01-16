package service

import (
	"context"
	"time"

	"github.com/ArtemMoroz51/FinalProject/internal/game"
)

type LeaderboardEntry struct {
	Place    int    `json:"place"`
	PlayerID string `json:"playerId"`
	Name     string `json:"name"`
	Score    int    `json:"score"`
}

type GameOverPayload struct {
	Code         string             `json:"code"`
	RoundsPlayed int                `json:"roundsPlayed"`
	Leaderboard  []LeaderboardEntry `json:"leaderboard"`
}

type Config struct {
	AnsweringSeconds time.Duration
	ResultsPause     time.Duration
	MaxRounds        int
}

type GameService interface {
	CreateRoom() *game.Room
	GetRoom(code string) (*game.Room, bool)

	StartRound(ctx context.Context, room *game.Room, hostID string) error

	MaxRounds() int
	AnsweringSeconds() time.Duration
	ResultsPause() time.Duration

	BuildLeaderboard(room *game.Room) GameOverPayload
}
