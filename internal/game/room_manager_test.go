package game

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func newTestRoomWithHost(t *testing.T) (*Room, *Player) {
	t.Helper()

	r := &Room{
		Code:    "ABCD",
		Phase:   PhaseLobby,
		Players: make(map[string]*Player),
		Answers: make(map[string]string),
		Scores:  make(map[string]int),
	}
	host := &Player{ID: "p1", Name: "Host"}
	isHost := r.AddPlayer(host)
	require.True(t, isHost)
	require.Equal(t, host.ID, r.HostID)
	return r, host
}

func validQuestion() Question {
	return Question{
		Text: "Q?",
		Options: []Option{
			{ID: "A", Text: "A"},
			{ID: "B", Text: "B"},
			{ID: "C", Text: "C"},
			{ID: "D", Text: "D"},
		},
		CorrectID: "B",
	}
}

func TestRoomManager_CreateRoom_Success(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom()

	require.NotNil(t, room)
	require.Len(t, room.Code, 4)
	require.Equal(t, PhaseLobby, room.Phase)
	require.NotNil(t, room.Players)
	require.NotNil(t, room.Answers)
	require.NotNil(t, room.Scores)

	got, ok := rm.GetRoom(room.Code)
	require.True(t, ok)
	require.Equal(t, room.Code, got.Code)
}

func TestRoomManager_GetRoom_CaseInsensitive(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom()

	got1, ok1 := rm.GetRoom(room.Code)
	require.True(t, ok1)
	require.Equal(t, room.Code, got1.Code)

	got2, ok2 := rm.GetRoom("  " + room.Code + "  ")
	require.False(t, ok2)
	require.Nil(t, got2)

	got3, ok3 := rm.GetRoom(stringsToLower(room.Code))
	require.True(t, ok3)
	require.Equal(t, room.Code, got3.Code)
}

func stringsToLower(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] = b[i] - 'A' + 'a'
		}
	}
	return string(b)
}

func TestRoom_StartGame_Success(t *testing.T) {
	r, host := newTestRoomWithHost(t)
	q := validQuestion()

	err := r.StartGame(host.ID, q, 30)
	require.NoError(t, err)

	snap := r.Snapshot()
	require.Equal(t, PhaseAnswering, snap.Phase)
	require.Equal(t, 1, snap.RoundNumber)
	require.Equal(t, q.Text, snap.Question)
	require.Len(t, snap.Options, 4)
	require.NotZero(t, snap.Deadline)
}

func TestRoom_StartGame_NotHost(t *testing.T) {
	r, _ := newTestRoomWithHost(t)
	q := validQuestion()

	err := r.StartGame("someone_else", q, 30)
	require.ErrorIs(t, err, ErrNotHost)
}

func TestRoom_StartGame_BadPhase(t *testing.T) {
	r, host := newTestRoomWithHost(t)
	r.Phase = PhaseAnswering
	q := validQuestion()

	err := r.StartGame(host.ID, q, 30)
	require.ErrorIs(t, err, ErrBadPhase)
}

func TestRoom_StartGame_InvalidQuestion(t *testing.T) {
	r, host := newTestRoomWithHost(t)

	err := r.StartGame(host.ID, Question{}, 30)
	require.ErrorIs(t, err, ErrInvalidQuestion)

	q := validQuestion()
	q.CorrectID = "Z"
	err = r.StartGame(host.ID, q, 30)
	require.ErrorIs(t, err, ErrInvalidQuestion)

	q2 := validQuestion()
	q2.Options = q2.Options[:3]
	err = r.StartGame(host.ID, q2, 30)
	require.ErrorIs(t, err, ErrInvalidQuestion)
}

func TestRoom_SubmitAnswer_Success(t *testing.T) {
	r, host := newTestRoomWithHost(t)
	q := validQuestion()
	require.NoError(t, r.StartGame(host.ID, q, 30))

	err := r.SubmitAnswer(host.ID, "A")
	require.NoError(t, err)

	require.Equal(t, "A", r.Answers[host.ID])
}

func TestRoom_SubmitAnswer_BadPhase(t *testing.T) {
	r, host := newTestRoomWithHost(t)

	err := r.SubmitAnswer(host.ID, "A")
	require.ErrorIs(t, err, ErrBadPhase)
}

func TestRoom_SubmitAnswer_Empty(t *testing.T) {
	r, host := newTestRoomWithHost(t)
	q := validQuestion()
	require.NoError(t, r.StartGame(host.ID, q, 30))

	err := r.SubmitAnswer(host.ID, "   ")
	require.ErrorIs(t, err, ErrEmptyAnswer)
}

func TestRoom_SubmitAnswer_InvalidOption(t *testing.T) {
	r, host := newTestRoomWithHost(t)
	q := validQuestion()
	require.NoError(t, r.StartGame(host.ID, q, 30))

	err := r.SubmitAnswer(host.ID, "Z")
	require.ErrorIs(t, err, ErrInvalidOption)
}

func TestRoom_SubmitAnswer_AlreadyAnswered(t *testing.T) {
	r, host := newTestRoomWithHost(t)
	q := validQuestion()
	require.NoError(t, r.StartGame(host.ID, q, 30))

	require.NoError(t, r.SubmitAnswer(host.ID, "A"))
	err := r.SubmitAnswer(host.ID, "B")
	require.ErrorIs(t, err, ErrAlreadyAnswered)
}

func TestRoom_SubmitAnswer_DeadlinePassed(t *testing.T) {
	r, host := newTestRoomWithHost(t)
	q := validQuestion()
	require.NoError(t, r.StartGame(host.ID, q, 1))

	r.AnsweringDeadline = time.Now().Add(-1 * time.Second)

	err := r.SubmitAnswer(host.ID, "A")
	require.ErrorIs(t, err, ErrDeadlinePassed)
}

func TestRoom_FinishRoundIfDeadlinePassed_NotYet(t *testing.T) {
	r, host := newTestRoomWithHost(t)
	q := validQuestion()
	require.NoError(t, r.StartGame(host.ID, q, 30))

	payload, ok := r.FinishRoundIfDeadlinePassed()
	require.False(t, ok)
	require.Nil(t, payload)
	require.Equal(t, PhaseAnswering, r.Phase)
}

func TestRoom_FinishRoundIfDeadlinePassed_SuccessAndScoring(t *testing.T) {
	r, host := newTestRoomWithHost(t)
	p2 := &Player{ID: "p2", Name: "P2"}
	r.AddPlayer(p2)

	q := validQuestion()
	require.NoError(t, r.StartGame(host.ID, q, 30))

	require.NoError(t, r.SubmitAnswer(host.ID, "B"))
	require.NoError(t, r.SubmitAnswer(p2.ID, "A"))

	r.AnsweringDeadline = time.Now().Add(-1 * time.Second)

	payload, ok := r.FinishRoundIfDeadlinePassed()
	require.True(t, ok)
	require.NotNil(t, payload)
	require.Equal(t, PhaseResults, r.Phase)
	require.Equal(t, 1, r.Scores[host.ID])
	require.Equal(t, 0, r.Scores[p2.ID])

	require.Equal(t, q.CorrectID, payload.CorrectOptionID)
	require.Len(t, payload.Results, 2)
}

func TestRoom_Snapshot_ScoresCopy(t *testing.T) {
	r, host := newTestRoomWithHost(t)
	q := validQuestion()
	require.NoError(t, r.StartGame(host.ID, q, 30))

	snap := r.Snapshot()
	require.NotNil(t, snap.Scores)

	snap.Scores[host.ID] = 999

	snap2 := r.Snapshot()
	require.NotEqual(t, 999, snap2.Scores[host.ID])
}
