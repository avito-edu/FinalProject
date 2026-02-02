package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ArtemMoroz51/FinalProject/internal/game"
	"github.com/ArtemMoroz51/FinalProject/internal/storage"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func makeRoomWithPlayers(t *testing.T) (*game.Room, *game.Player, *game.Player) {
	t.Helper()

	r := &game.Room{
		Code:    "ABCD",
		Phase:   game.PhaseLobby,
		Players: make(map[string]*game.Player),
		Answers: make(map[string]string),
		Scores:  make(map[string]int),
	}
	host := &game.Player{ID: "p1", Name: "Host"}
	p2 := &game.Player{ID: "p2", Name: "Alice"}
	r.AddPlayer(host)
	r.AddPlayer(p2)
	return r, host, p2
}

func validQuestion() game.Question {
	return game.Question{
		Text: "Q?",
		Options: []game.Option{
			{ID: "A", Text: "A"},
			{ID: "B", Text: "B"},
			{ID: "C", Text: "C"},
			{ID: "D", Text: "D"},
		},
		CorrectID: "B",
	}
}

func TestNewGameService_Defaults(t *testing.T) {
	rm := game.NewRoomManager()
	qs := new(mockQuestionStore)

	svc := NewGameService(rm, qs, Config{})
	require.Equal(t, 5, svc.MaxRounds())
	require.Equal(t, 30*time.Second, svc.AnsweringSeconds())
	require.Equal(t, 5*time.Second, svc.ResultsPause())
}

func TestGameService_StartRound_Success(t *testing.T) {
	rm := game.NewRoomManager()
	qs := new(mockQuestionStore)

	cfg := Config{
		AnsweringSeconds: 10 * time.Second,
		ResultsPause:     3 * time.Second,
		MaxRounds:        7,
	}
	svc := NewGameService(rm, qs, cfg)

	room, host, _ := makeRoomWithPlayers(t)
	q := validQuestion()

	qs.On("GetRandomActive", mock.Anything).Return(q, nil).Once()

	err := svc.StartRound(context.Background(), room, host.ID)
	require.NoError(t, err)

	snap := room.Snapshot()
	require.Equal(t, game.PhaseAnswering, snap.Phase)
	require.Equal(t, 1, snap.RoundNumber)
	require.Equal(t, q.Text, snap.Question)
	require.NotZero(t, snap.Deadline)

	qs.AssertExpectations(t)
}

func TestGameService_StartRound_NoQuestions(t *testing.T) {
	rm := game.NewRoomManager()
	qs := new(mockQuestionStore)

	svc := NewGameService(rm, qs, Config{})

	room, host, _ := makeRoomWithPlayers(t)

	qs.On("GetRandomActive", mock.Anything).Return(game.Question{}, storage.ErrNoQuestions).Once()

	err := svc.StartRound(context.Background(), room, host.ID)
	require.Error(t, err)
	require.Equal(t, "no questions in db", err.Error())

	qs.AssertExpectations(t)
}

func TestGameService_StartRound_RepoError_Passthrough(t *testing.T) {
	rm := game.NewRoomManager()
	qs := new(mockQuestionStore)

	svc := NewGameService(rm, qs, Config{})

	room, host, _ := makeRoomWithPlayers(t)

	repoErr := errors.New("db down")
	qs.On("GetRandomActive", mock.Anything).Return(game.Question{}, repoErr).Once()

	err := svc.StartRound(context.Background(), room, host.ID)
	require.ErrorIs(t, err, repoErr)

	qs.AssertExpectations(t)
}

func TestGameService_BuildLeaderboard_SortsAndPlaces(t *testing.T) {
	rm := game.NewRoomManager()
	qs := new(mockQuestionStore)
	svc := NewGameService(rm, qs, Config{})

	room, host, p2 := makeRoomWithPlayers(t)

	room.Scores[host.ID] = 2
	room.Scores[p2.ID] = 2
	host.Name = "Bob"
	p2.Name = "Alice"

	payload := svc.BuildLeaderboard(room)

	require.Equal(t, room.Code, payload.Code)
	require.Equal(t, room.RoundNumber, payload.RoundsPlayed)
	require.Len(t, payload.Leaderboard, 2)

	require.Equal(t, "Alice", payload.Leaderboard[0].Name)
	require.Equal(t, 1, payload.Leaderboard[0].Place)
	require.Equal(t, 2, payload.Leaderboard[0].Score)

	require.Equal(t, "Bob", payload.Leaderboard[1].Name)
	require.Equal(t, 1, payload.Leaderboard[1].Place)
	require.Equal(t, 2, payload.Leaderboard[1].Score)
}
