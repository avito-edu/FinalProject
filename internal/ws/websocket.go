package ws

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/ArtemMoroz51/FinalProject/internal/game"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request, roomCode string) {
	room, ok := h.svc.GetRoom(roomCode)
	if !ok {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	_ = conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	var msg clientMsg
	if err := conn.ReadJSON(&msg); err != nil || msg.Type != "join_room" {
		_ = conn.WriteJSON(Envelope{Type: "error", Payload: map[string]string{"message": "expected join_room"}})
		_ = conn.Close()
		return
	}

	var jp JoinPayload
	if err := json.Unmarshal(msg.Payload, &jp); err != nil || strings.TrimSpace(jp.Name) == "" {
		_ = conn.WriteJSON(Envelope{Type: "error", Payload: map[string]string{"message": "invalid name"}})
		_ = conn.Close()
		return
	}
	_ = conn.SetReadDeadline(time.Time{})

	playerID := newID()
	player := &game.Player{ID: playerID, Name: strings.TrimSpace(jp.Name)}
	_ = room.AddPlayer(player)

	client := &Client{
		hub:      h,
		roomCode: strings.ToUpper(roomCode),
		playerID: playerID,
		conn:     conn,
		send:     make(chan []byte, 64),
	}

	h.register <- client
	go client.writePump()

	h.Broadcast(roomCode, Envelope{Type: "player_joined", Payload: player})
	h.Broadcast(roomCode, Envelope{Type: "room_state", Payload: room.Snapshot()})

	client.readPump(room)
}
