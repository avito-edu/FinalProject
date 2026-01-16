package ws

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ArtemMoroz51/FinalProject/internal/game"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Client struct {
	hub      *Hub
	roomCode string
	playerID string
	conn     *websocket.Conn
	send     chan []byte
}

func (c *Client) sendJSON(env Envelope) {
	b, err := json.Marshal(env)
	if err != nil {
		c.hub.log.Error("ws send marshal failed",
			zap.String("room", c.roomCode),
			zap.String("player_id", c.playerID),
			zap.Error(err),
		)
		return
	}
	select {
	case c.send <- b:
	default:
		c.hub.unregister <- c
	}
}

func (c *Client) readPump(room *game.Room) {
	defer func() {
		room.RemovePlayer(c.playerID)
		c.hub.Broadcast(c.roomCode, Envelope{Type: "room_state", Payload: room.Snapshot()})
		c.hub.unregister <- c
		_ = c.conn.Close()

		c.hub.log.Info("ws connection closed",
			zap.String("room", c.roomCode),
			zap.String("player_id", c.playerID),
		)
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var msg clientMsg
		if err := c.conn.ReadJSON(&msg); err != nil {
			c.hub.log.Warn("ws read failed",
				zap.String("room", c.roomCode),
				zap.String("player_id", c.playerID),
				zap.Error(err),
			)
			break
		}

		c.hub.log.Info("ws message received",
			zap.String("room", c.roomCode),
			zap.String("player_id", c.playerID),
			zap.String("type", msg.Type),
		)

		switch msg.Type {
		case "start_game":
			snap := room.Snapshot()
			if snap.RoundNumber >= c.hub.svc.MaxRounds() {
				gameOver := c.hub.svc.BuildLeaderboard(room)
				c.sendJSON(Envelope{Type: "game_over", Payload: gameOver})
				continue
			}

			if err := c.hub.svc.StartRound(context.Background(), room, c.playerID); err != nil {
				c.hub.log.Warn("start_game failed",
					zap.String("room", c.roomCode),
					zap.String("player_id", c.playerID),
					zap.Error(err),
				)
				c.sendJSON(Envelope{Type: "error", Payload: map[string]string{"message": err.Error()}})
				continue
			}

			c.hub.Broadcast(c.roomCode, Envelope{Type: "room_state", Payload: room.Snapshot()})

			gen := c.hub.bumpRoundGen(c.roomCode)
			go c.hub.scheduleAnsweringDeadline(room, c.roomCode, gen)

		case "submit_answer":
			var p SubmitAnswerPayload
			if err := json.Unmarshal(msg.Payload, &p); err != nil {
				c.hub.log.Warn("submit_answer bad payload",
					zap.String("room", c.roomCode),
					zap.String("player_id", c.playerID),
					zap.Error(err),
				)
				c.sendJSON(Envelope{Type: "error", Payload: map[string]string{"message": "bad payload"}})
				continue
			}

			if err := room.SubmitAnswer(c.playerID, p.OptionID); err != nil {
				c.hub.log.Warn("submit_answer failed",
					zap.String("room", c.roomCode),
					zap.String("player_id", c.playerID),
					zap.String("option_id", p.OptionID),
					zap.Error(err),
				)
				c.sendJSON(Envelope{Type: "error", Payload: map[string]string{"message": err.Error()}})
				continue
			}

			c.sendJSON(Envelope{Type: "answer_accepted", Payload: map[string]bool{"ok": true}})

		default:
			c.hub.log.Warn("unknown ws message type",
				zap.String("room", c.roomCode),
				zap.String("player_id", c.playerID),
				zap.String("type", msg.Type),
			)
			c.sendJSON(Envelope{Type: "error", Payload: map[string]string{"message": "unknown message type"}})
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.hub.log.Warn("ws write failed",
					zap.String("room", c.roomCode),
					zap.String("player_id", c.playerID),
					zap.Error(err),
				)
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.hub.log.Warn("ws ping failed",
					zap.String("room", c.roomCode),
					zap.String("player_id", c.playerID),
					zap.Error(err),
				)
				return
			}
		}
	}
}
