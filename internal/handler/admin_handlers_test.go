package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ArtemMoroz51/FinalProject/internal/game"
	"github.com/ArtemMoroz51/FinalProject/internal/storage"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockAdminService struct {
	mock.Mock
}

func (m *mockAdminService) CreateQuestion(ctx context.Context, in storage.CreateQuestionInput) (storage.QuestionRow, error) {
	args := m.Called(ctx, in)
	row, _ := args.Get(0).(storage.QuestionRow)
	return row, args.Error(1)
}

func (m *mockAdminService) ListQuestions(ctx context.Context, includeInactive bool) ([]storage.QuestionRow, error) {
	args := m.Called(ctx, includeInactive)
	rows, _ := args.Get(0).([]storage.QuestionRow)
	return rows, args.Error(1)
}

func (m *mockAdminService) SetQuestionActive(ctx context.Context, id int64, active bool) (storage.QuestionRow, error) {
	args := m.Called(ctx, id, active)
	row, _ := args.Get(0).(storage.QuestionRow)
	return row, args.Error(1)
}

func TestAdminHandlers_PostQuestions_BadJSON(t *testing.T) {
	mux := http.NewServeMux()
	admin := new(mockAdminService)

	RegisterAdminHandlers(mux, admin, "token123", zap.NewNop())

	req := httptest.NewRequest(http.MethodPost, "/admin/questions", bytes.NewBufferString("{bad json"))
	req.Header.Set("Authorization", "Bearer token123")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "bad json")
	admin.AssertExpectations(t)
}

func TestAdminHandlers_PatchQuestions_BadID(t *testing.T) {
	mux := http.NewServeMux()
	admin := new(mockAdminService)

	RegisterAdminHandlers(mux, admin, "token123", zap.NewNop())

	body, _ := json.Marshal(map[string]bool{"isActive": true})
	req := httptest.NewRequest(http.MethodPatch, "/admin/questions/abc", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer token123")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "bad id")
	admin.AssertExpectations(t)
}

func TestAdminHandlers_GetQuestions_ServiceError(t *testing.T) {
	mux := http.NewServeMux()
	admin := new(mockAdminService)

	RegisterAdminHandlers(mux, admin, "token123", zap.NewNop())

	admin.On("ListQuestions", mock.Anything, false).Return([]storage.QuestionRow(nil), errors.New("boom")).Once()

	req := httptest.NewRequest(http.MethodGet, "/admin/questions", nil)
	req.Header.Set("Authorization", "Bearer token123")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	admin.AssertExpectations(t)
}

func TestAdminHandlers_PostQuestions_Success(t *testing.T) {
	mux := http.NewServeMux()
	admin := new(mockAdminService)

	RegisterAdminHandlers(mux, admin, "token123", zap.NewNop())

	in := storage.CreateQuestionInput{
		Text: "Q",
		Options: []game.Option{
			{ID: "A", Text: "a"},
			{ID: "B", Text: "b"},
			{ID: "C", Text: "c"},
			{ID: "D", Text: "d"},
		},
		CorrectID: "A",
		IsActive:  true,
	}

	b, _ := json.Marshal(in)

	expectedRow := storage.QuestionRow{ID: 1, Text: "Q", CorrectID: "A", IsActive: true}
	admin.On("CreateQuestion", mock.Anything, mock.Anything).Return(expectedRow, nil).Once()

	req := httptest.NewRequest(http.MethodPost, "/admin/questions", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer token123")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"id":1`)
	admin.AssertExpectations(t)
}
