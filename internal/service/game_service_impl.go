package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/ArtemMoroz51/FinalProject/internal/game"
	"github.com/ArtemMoroz51/FinalProject/internal/storage"
)

type gameService struct {
	rm  *game.RoomManager
	qs  storage.QuestionStore
	cfg Config
}

func NewGameService(rm *game.RoomManager, qs storage.QuestionStore, cfg Config) GameService {
	if cfg.AnsweringSeconds == 0 {
		cfg.AnsweringSeconds = 30 * time.Second
	}
	if cfg.ResultsPause == 0 {
		cfg.ResultsPause = 5 * time.Second
	}
	if cfg.MaxRounds == 0 {
		cfg.MaxRounds = 5
	}
	return &gameService{rm: rm, qs: qs, cfg: cfg}
}

func (s *gameService) CreateRoom() *game.Room {
	return s.rm.CreateRoom()
}

func (s *gameService) GetRoom(code string) (*game.Room, bool) {
	return s.rm.GetRoom(code)
}

func (s *gameService) StartRound(ctx context.Context, room *game.Room, hostID string) error {
	q, err := s.qs.GetRandomActive(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrNoQuestions) {
			return fmt.Errorf("no questions in db")
		}
		return err
	}
	return room.StartGame(hostID, q, int(s.cfg.AnsweringSeconds.Seconds()))
}

func (s *gameService) MaxRounds() int                  { return s.cfg.MaxRounds }
func (s *gameService) AnsweringSeconds() time.Duration { return s.cfg.AnsweringSeconds }
func (s *gameService) ResultsPause() time.Duration     { return s.cfg.ResultsPause }

func (s *gameService) BuildLeaderboard(room *game.Room) GameOverPayload {
	snap := room.Snapshot()

	type row struct {
		id    string
		name  string
		score int
	}
	rows := make([]row, 0, len(snap.Players))
	for _, p := range snap.Players {
		rows = append(rows, row{
			id:    p.ID,
			name:  p.Name,
			score: snap.Scores[p.ID],
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].score != rows[j].score {
			return rows[i].score > rows[j].score
		}
		return rows[i].name < rows[j].name
	})

	leaderboard := make([]LeaderboardEntry, 0, len(rows))
	place := 0
	prevScore := -1
	for i, r := range rows {
		if i == 0 || r.score != prevScore {
			place = i + 1
			prevScore = r.score
		}
		leaderboard = append(leaderboard, LeaderboardEntry{
			Place:    place,
			PlayerID: r.id,
			Name:     r.name,
			Score:    r.score,
		})
	}

	return GameOverPayload{
		Code:         snap.Code,
		RoundsPlayed: snap.RoundNumber,
		Leaderboard:  leaderboard,
	}
}
