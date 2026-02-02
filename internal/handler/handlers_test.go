package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ArtemMoroz51/FinalProject/internal/game"
	"github.com/ArtemMoroz51/FinalProject/internal/service"
	"github.com/ArtemMoroz51/FinalProject/internal/ws"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockGameService struct {
	mock.Mock
}

func (m *mockGameService) CreateRoom() *game.Room {
	args := m.Called()
	r, _ := args.Get(0).(*game.Room)
	return r
}

func (m *mockGameService) GetRoom(code string) (*game.Room, bool) {
	args := m.Called(code)
	r, _ := args.Get(0).(*game.Room)
	ok, _ := args.Get(1).(bool)
	return r, ok
}

func (m *mockGameService) StartRound(ctx context.Context, room *game.Room, hostID string) error {
	args := m.Called(ctx, room, hostID)
	return args.Error(0)
}

func (m *mockGameService) MaxRounds() int {
	args := m.Called()
	return args.Int(0)
}

func (m *mockGameService) AnsweringSeconds() time.Duration {
	args := m.Called()
	d, _ := args.Get(0).(time.Duration)
	return d
}

func (m *mockGameService) ResultsPause() time.Duration {
	args := m.Called()
	d, _ := args.Get(0).(time.Duration)
	return d
}

func (m *mockGameService) BuildLeaderboard(room *game.Room) service.GameOverPayload {
	args := m.Called(room)
	p, _ := args.Get(0).(service.GameOverPayload)
	return p
}

func TestHandlers_PostRooms_MethodNotAllowed(t *testing.T) {
	mux := http.NewServeMux()
	svc := new(mockGameService)
	hub := ws.NewHub(nil, zap.NewNop())
	RegisterHandlers(mux, svc, hub, zap.NewNop())

	req := httptest.NewRequest(http.MethodGet, "/rooms", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	require.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandlers_PostRooms_Success(t *testing.T) {
	mux := http.NewServeMux()
	svc := new(mockGameService)

	room := &game.Room{Code: "ABCD", Phase: game.PhaseLobby}
	svc.On("CreateRoom").Return(room).Once()

	hub := ws.NewHub(nil, zap.NewNop())
	RegisterHandlers(mux, svc, hub, zap.NewNop())

	req := httptest.NewRequest(http.MethodPost, "/rooms", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "ABCD", resp["code"])

	svc.AssertExpectations(t)
}

func TestHandlers_GetRoom_MethodNotAllowed(t *testing.T) {
	mux := http.NewServeMux()
	svc := new(mockGameService)
	hub := ws.NewHub(nil, zap.NewNop())
	RegisterHandlers(mux, svc, hub, zap.NewNop())

	req := httptest.NewRequest(http.MethodPost, "/rooms/ABCD", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	require.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandlers_GetRoom_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	svc := new(mockGameService)

	svc.On("GetRoom", "ABCD").Return((*game.Room)(nil), false).Once()

	hub := ws.NewHub(nil, zap.NewNop())
	RegisterHandlers(mux, svc, hub, zap.NewNop())

	req := httptest.NewRequest(http.MethodGet, "/rooms/ABCD", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

func TestHandlers_GetRoom_Success(t *testing.T) {
	mux := http.NewServeMux()
	svc := new(mockGameService)

	room := &game.Room{Code: "ABCD", Phase: game.PhaseLobby}
	svc.On("GetRoom", "ABCD").Return(room, true).Once()

	hub := ws.NewHub(nil, zap.NewNop())
	RegisterHandlers(mux, svc, hub, zap.NewNop())

	req := httptest.NewRequest(http.MethodGet, "/rooms/ABCD", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "ABCD", resp["code"])
	require.NotNil(t, resp["phase"])

	svc.AssertExpectations(t)
}
