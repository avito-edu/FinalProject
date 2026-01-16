package ws

import (
	"encoding/json"
	"sort"
	"strings"
	"sync"

	"github.com/ArtemMoroz51/FinalProject/internal/game"
	"github.com/ArtemMoroz51/FinalProject/internal/service"
	"go.uber.org/zap"
)

type Hub struct {
	svc service.GameService
	log *zap.Logger

	mu            sync.RWMutex
	clientsByRoom map[string]map[string]*Client

	register   chan *Client
	unregister chan *Client
	broadcast  chan roomMessage

	roundGenMu sync.Mutex
	roundGen   map[string]int64
}

type roomMessage struct {
	roomCode string
	data     []byte
}

func NewHub(svc service.GameService, log *zap.Logger) *Hub {
	if log == nil {
		log = zap.NewNop()
	}
	h := &Hub{
		svc:           svc,
		log:           log,
		clientsByRoom: make(map[string]map[string]*Client),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		broadcast:     make(chan roomMessage, 256),
		roundGen:      make(map[string]int64),
	}
	go h.run()
	return h
}

func (h *Hub) Broadcast(roomCode string, env Envelope) {
	b, err := json.Marshal(env)
	if err != nil {
		h.log.Error("ws broadcast marshal failed", zap.Error(err))
		return
	}
	h.broadcast <- roomMessage{roomCode: roomCode, data: b}
}

func (h *Hub) run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			roomCode := strings.ToUpper(c.roomCode)
			if _, ok := h.clientsByRoom[roomCode]; !ok {
				h.clientsByRoom[roomCode] = make(map[string]*Client)
			}
			h.clientsByRoom[roomCode][c.playerID] = c
			h.mu.Unlock()

			h.log.Info("ws client registered",
				zap.String("room", roomCode),
				zap.String("player_id", c.playerID),
			)

		case c := <-h.unregister:
			h.mu.Lock()
			roomCode := strings.ToUpper(c.roomCode)
			if roomClients, ok := h.clientsByRoom[roomCode]; ok {
				if _, exists := roomClients[c.playerID]; exists {
					delete(roomClients, c.playerID)
					close(c.send)
				}
				if len(roomClients) == 0 {
					delete(h.clientsByRoom, roomCode)
				}
			}
			h.mu.Unlock()

			h.log.Info("ws client unregistered",
				zap.String("room", roomCode),
				zap.String("player_id", c.playerID),
			)

		case msg := <-h.broadcast:
			h.mu.RLock()
			roomClients := h.clientsByRoom[strings.ToUpper(msg.roomCode)]
			for _, c := range roomClients {
				select {
				case c.send <- msg.data:
				default:
					h.mu.RUnlock()
					h.unregister <- c
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) bumpRoundGen(roomCode string) int64 {
	h.roundGenMu.Lock()
	defer h.roundGenMu.Unlock()
	rc := strings.ToUpper(roomCode)
	h.roundGen[rc]++
	return h.roundGen[rc]
}

func (h *Hub) isCurrentGen(roomCode string, gen int64) bool {
	h.roundGenMu.Lock()
	defer h.roundGenMu.Unlock()
	rc := strings.ToUpper(roomCode)
	return h.roundGen[rc] == gen
}

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

func buildLeaderboard(room *game.Room) GameOverPayload {
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
