package ws

import (
	"context"
	"time"

	"github.com/ArtemMoroz51/FinalProject/internal/game"
)

func (h *Hub) scheduleAnsweringDeadline(room *game.Room, roomCode string, gen int64) {
	snap := room.Snapshot()
	if snap.Deadline == 0 {
		return
	}

	wait := time.Until(time.UnixMilli(snap.Deadline))
	if wait < 0 {
		wait = 0
	}
	time.Sleep(wait)

	if !h.isCurrentGen(roomCode, gen) {
		return
	}

	if payload, ok := room.FinishRoundIfDeadlinePassed(); ok {
		h.Broadcast(roomCode, Envelope{Type: "round_results", Payload: payload})
		h.Broadcast(roomCode, Envelope{Type: "room_state", Payload: room.Snapshot()})

		after := room.Snapshot()
		if after.RoundNumber >= h.svc.MaxRounds() {
			gameOver := h.svc.BuildLeaderboard(room)
			h.Broadcast(roomCode, Envelope{Type: "game_over", Payload: gameOver})
			return
		}

		go h.scheduleNextRound(room, roomCode, h.svc.ResultsPause())
	}
}

func (h *Hub) scheduleNextRound(room *game.Room, roomCode string, delay time.Duration) {
	time.Sleep(delay)

	snap := room.Snapshot()
	if snap.HostID == "" {
		return
	}
	if snap.RoundNumber >= h.svc.MaxRounds() {
		gameOver := h.svc.BuildLeaderboard(room)
		h.Broadcast(roomCode, Envelope{Type: "game_over", Payload: gameOver})
		return
	}

	if err := h.svc.StartRound(context.Background(), room, snap.HostID); err != nil {
		h.Broadcast(roomCode, Envelope{Type: "error", Payload: map[string]string{"message": err.Error()}})
		return
	}

	h.Broadcast(roomCode, Envelope{Type: "room_state", Payload: room.Snapshot()})

	gen := h.bumpRoundGen(roomCode)
	go h.scheduleAnsweringDeadline(room, roomCode, gen)
}
