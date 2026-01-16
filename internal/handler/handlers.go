package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ArtemMoroz51/FinalProject/internal/service"
	"github.com/ArtemMoroz51/FinalProject/internal/ws"
	"go.uber.org/zap"
)

func RegisterHandlers(mux *http.ServeMux, svc service.GameService, hub *ws.Hub, log *zap.Logger) {
	if log == nil {
		log = zap.NewNop()
	}

	mux.HandleFunc("/rooms", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.Warn("method not allowed", zap.String("path", r.URL.Path), zap.String("method", r.Method))
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		room := svc.CreateRoom()
		log.Info("room created", zap.String("code", room.Code))
		_ = json.NewEncoder(w).Encode(map[string]string{"code": room.Code})
	})

	mux.HandleFunc("/rooms/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			log.Warn("method not allowed", zap.String("path", r.URL.Path), zap.String("method", r.Method))
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		code := strings.TrimPrefix(r.URL.Path, "/rooms/")
		room, ok := svc.GetRoom(code)
		if !ok {
			log.Warn("room not found", zap.String("code", code))
			http.Error(w, "room not found", http.StatusNotFound)
			return
		}
		log.Info("room fetched", zap.String("code", room.Code))
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":  room.Code,
			"phase": room.Phase,
		})
	})

	mux.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
		code := strings.TrimPrefix(r.URL.Path, "/ws/")
		log.Info("ws connect attempt", zap.String("code", code))
		hub.ServeWS(w, r, code)
	})
}
