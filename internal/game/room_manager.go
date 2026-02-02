package game

import (
	"crypto/rand"
	"encoding/base32"
	"strings"
	"sync"
	"time"
)

type Option struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type Question struct {
	Text      string   `json:"text"`
	Options   []Option `json:"options"`
	CorrectID string   `json:"-"`
}

type Room struct {
	Code    string
	Phase   Phase
	Players map[string]*Player

	HostID string

	RoundNumber     int
	CurrentQuestion Question

	AnsweringDeadline time.Time

	Answers map[string]string
	Scores  map[string]int

	mu sync.Mutex
}

type RoomSnapshot struct {
	Code        string `json:"code"`
	Phase       Phase  `json:"phase"`
	HostID      string `json:"hostId"`
	RoundNumber int    `json:"roundNumber"`

	Question string   `json:"question,omitempty"`
	Options  []Option `json:"options,omitempty"`

	Deadline int64          `json:"deadline,omitempty"`
	Players  []*Player      `json:"players"`
	Scores   map[string]int `json:"scores"`
}

func (r *Room) AddPlayer(p *Player) (isHost bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Players[p.ID] = p

	if r.Scores != nil {
		if _, ok := r.Scores[p.ID]; !ok {
			r.Scores[p.ID] = 0
		}
	}

	if r.HostID == "" {
		r.HostID = p.ID
		return true
	}
	return false
}

func (r *Room) RemovePlayer(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.Players, playerID)

	if r.HostID == playerID {
		r.HostID = ""
		for id := range r.Players {
			r.HostID = id
			break
		}
	}
}


func (r *Room) StartGame(requesterID string, q Question, answeringSeconds int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.HostID == "" || r.HostID != requesterID {
		return ErrNotHost
	}
	if r.Phase != PhaseLobby && r.Phase != PhaseResults {
		return ErrBadPhase
	}
	if len(r.Players) < 1 {
		return ErrNoPlayers
	}

	if strings.TrimSpace(q.Text) == "" || len(q.Options) != 4 || strings.TrimSpace(q.CorrectID) == "" {
		return ErrInvalidQuestion
	}
	if !hasOption(q.Options, q.CorrectID) {
		return ErrInvalidQuestion
	}

	r.RoundNumber++
	r.CurrentQuestion = q

	r.Answers = make(map[string]string)

	if r.Scores == nil {
		r.Scores = make(map[string]int)
	}
	for id := range r.Players {
		if _, ok := r.Scores[id]; !ok {
			r.Scores[id] = 0
		}
	}

	r.Phase = PhaseAnswering
	r.AnsweringDeadline = time.Now().Add(time.Duration(answeringSeconds) * time.Second)
	return nil
}

func (r *Room) SubmitAnswer(playerID string, optionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Phase != PhaseAnswering {
		return ErrBadPhase
	}
	if !r.AnsweringDeadline.IsZero() && time.Now().After(r.AnsweringDeadline) {
		return ErrDeadlinePassed
	}

	optionID = strings.TrimSpace(optionID)
	if optionID == "" {
		return ErrEmptyAnswer
	}
	if !hasOption(r.CurrentQuestion.Options, optionID) {
		return ErrInvalidOption
	}

	if r.Answers == nil {
		r.Answers = make(map[string]string)
	}
	if _, ok := r.Answers[playerID]; ok {
		return ErrAlreadyAnswered
	}

	r.Answers[playerID] = optionID
	return nil
}

type RoundResult struct {
	PlayerID         string `json:"playerId"`
	Name             string `json:"name"`
	SelectedOptionID string `json:"selectedOptionId,omitempty"`
	Correct          bool   `json:"correct"`
	Score            int    `json:"score"`
}

type RoundResultsPayload struct {
	Code            string        `json:"code"`
	RoundNumber     int           `json:"roundNumber"`
	Question        string        `json:"question"`
	Options         []Option      `json:"options"`
	CorrectOptionID string        `json:"correctOptionId"`
	Results         []RoundResult `json:"results"`
}

func (r *Room) FinishRoundIfDeadlinePassed() (*RoundResultsPayload, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Phase != PhaseAnswering {
		return nil, false
	}
	if time.Now().Before(r.AnsweringDeadline) {
		return nil, false
	}

	if r.Scores == nil {
		r.Scores = make(map[string]int)
	}
	if r.Answers == nil {
		r.Answers = make(map[string]string)
	}

	correctID := r.CurrentQuestion.CorrectID

	results := make([]RoundResult, 0, len(r.Players))
	for id, p := range r.Players {
		selected := r.Answers[id]
		isCorrect := selected != "" && selected == correctID
		if isCorrect {
			r.Scores[id]++
		}

		results = append(results, RoundResult{
			PlayerID:         id,
			Name:             p.Name,
			SelectedOptionID: selected,
			Correct:          isCorrect,
			Score:            r.Scores[id],
		})
	}

	r.Phase = PhaseResults

	payload := &RoundResultsPayload{
		Code:            r.Code,
		RoundNumber:     r.RoundNumber,
		Question:        r.CurrentQuestion.Text,
		Options:         r.CurrentQuestion.Options,
		CorrectOptionID: correctID,
		Results:         results,
	}
	return payload, true
}

func (r *Room) Snapshot() RoomSnapshot {
	r.mu.Lock()
	defer r.mu.Unlock()

	players := make([]*Player, 0, len(r.Players))
	for _, p := range r.Players {
		players = append(players, p)
	}

	var deadlineMillis int64
	if r.Phase == PhaseAnswering && !r.AnsweringDeadline.IsZero() {
		deadlineMillis = r.AnsweringDeadline.UnixMilli()
	}

	scoresCopy := make(map[string]int)
	if r.Scores != nil {
		scoresCopy = make(map[string]int, len(r.Scores))
		for k, v := range r.Scores {
			scoresCopy[k] = v
		}
	}

	s := RoomSnapshot{
		Code:        r.Code,
		Phase:       r.Phase,
		HostID:      r.HostID,
		RoundNumber: r.RoundNumber,
		Players:     players,
		Scores:      scoresCopy,
	}

	if r.Phase == PhaseAnswering || r.Phase == PhaseResults {
		s.Question = r.CurrentQuestion.Text
		s.Options = r.CurrentQuestion.Options
	}

	if deadlineMillis != 0 {
		s.Deadline = deadlineMillis
	}

	return s
}

func hasOption(opts []Option, id string) bool {
	for _, o := range opts {
		if o.ID == id {
			return true
		}
	}
	return false
}

type RoomManager struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

func NewRoomManager() *RoomManager {
	return &RoomManager{rooms: make(map[string]*Room)}
}

func (rm *RoomManager) CreateRoom() *Room {
	code := rm.generateCode(4)
	room := &Room{
		Code:    code,
		Phase:   PhaseLobby,
		Players: make(map[string]*Player),
		Answers: make(map[string]string),
		Scores:  make(map[string]int),
	}

	rm.mu.Lock()
	rm.rooms[code] = room
	rm.mu.Unlock()

	return room
}

func (rm *RoomManager) GetRoom(code string) (*Room, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	r, ok := rm.rooms[strings.ToUpper(code)]
	return r, ok
}

func (rm *RoomManager) generateCode(n int) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)

	s := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
	s = strings.NewReplacer("O", "A", "I", "B", "0", "C", "1", "D").Replace(s)

	return strings.ToUpper(s[:n])
}
