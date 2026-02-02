package service

import (
	"context"
	"errors"
	"testing"

	"github.com/ArtemMoroz51/FinalProject/internal/game"
	"github.com/ArtemMoroz51/FinalProject/internal/storage"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockQuestionStore struct {
	mock.Mock
}

func (m *mockQuestionStore) CreateQuestion(ctx context.Context, in storage.CreateQuestionInput) (storage.QuestionRow, error) {
	args := m.Called(ctx, in)
	row, _ := args.Get(0).(storage.QuestionRow)
	return row, args.Error(1)
}

func (m *mockQuestionStore) ListQuestions(ctx context.Context, includeInactive bool) ([]storage.QuestionRow, error) {
	args := m.Called(ctx, includeInactive)
	rows, _ := args.Get(0).([]storage.QuestionRow)
	return rows, args.Error(1)
}

func (m *mockQuestionStore) SetQuestionActive(ctx context.Context, id int64, active bool) (storage.QuestionRow, error) {
	args := m.Called(ctx, id, active)
	row, _ := args.Get(0).(storage.QuestionRow)
	return row, args.Error(1)
}
func (m *mockQuestionStore) GetRandomActive(ctx context.Context) (game.Question, error) {
	args := m.Called(ctx)
	q, _ := args.Get(0).(game.Question)
	return q, args.Error(1)
}

func TestAdminService_CreateQuestion_InvalidPayload(t *testing.T) {
	qs := new(mockQuestionStore)
	svc := NewAdminService(qs)

	ctx := context.Background()

	_, err := svc.CreateQuestion(ctx, storage.CreateQuestionInput{})
	require.Error(t, err)

	_, err = svc.CreateQuestion(ctx, storage.CreateQuestionInput{
		Text:      "Q",
		CorrectID: "A",
		Options:   []game.Option{},
	})
	require.Error(t, err)

	qs.AssertNotCalled(t, "CreateQuestion", mock.Anything, mock.Anything)
}

func TestAdminService_CreateQuestion_Success_TrimsAndCallsRepo(t *testing.T) {
	qs := new(mockQuestionStore)
	svc := NewAdminService(qs)

	ctx := context.Background()

	in := storage.CreateQuestionInput{
		Text:      "  Q1  ",
		CorrectID: "  B ",
		Options: []game.Option{
			{ID: "A", Text: "a"},
			{ID: "B", Text: "b"},
			{ID: "C", Text: "c"},
			{ID: "D", Text: "d"},
		},
		IsActive: true,
	}

	expectedIn := in
	expectedIn.Text = "Q1"
	expectedIn.CorrectID = "B"

	expectedRow := storage.QuestionRow{ID: 10, Text: "Q1", CorrectID: "B", IsActive: true}

	qs.On("CreateQuestion", mock.Anything, expectedIn).Return(expectedRow, nil).Once()

	row, err := svc.CreateQuestion(ctx, in)
	require.NoError(t, err)
	require.Equal(t, expectedRow, row)

	qs.AssertExpectations(t)
}

func TestAdminService_CreateQuestion_RepoError(t *testing.T) {
	qs := new(mockQuestionStore)
	svc := NewAdminService(qs)

	ctx := context.Background()

	in := storage.CreateQuestionInput{
		Text:      "Q1",
		CorrectID: "B",
		Options: []game.Option{
			{ID: "A", Text: "a"},
			{ID: "B", Text: "b"},
			{ID: "C", Text: "c"},
			{ID: "D", Text: "d"},
		},
		IsActive: true,
	}

	repoErr := errors.New("db error")

	qs.On("CreateQuestion", mock.Anything, in).Return(storage.QuestionRow{}, repoErr).Once()

	_, err := svc.CreateQuestion(ctx, in)
	require.Error(t, err)
	require.Equal(t, repoErr, err)

	qs.AssertExpectations(t)
}

func TestAdminService_ListQuestions_Passthrough(t *testing.T) {
	qs := new(mockQuestionStore)
	svc := NewAdminService(qs)

	ctx := context.Background()

	expected := []storage.QuestionRow{
		{ID: 1, Text: "Q1", IsActive: true},
		{ID: 2, Text: "Q2", IsActive: false},
	}

	qs.On("ListQuestions", mock.Anything, true).Return(expected, nil).Once()

	rows, err := svc.ListQuestions(ctx, true)
	require.NoError(t, err)
	require.Equal(t, expected, rows)

	qs.AssertExpectations(t)
}

func TestAdminService_SetQuestionActive_InvalidID(t *testing.T) {
	qs := new(mockQuestionStore)
	svc := NewAdminService(qs)

	ctx := context.Background()

	_, err := svc.SetQuestionActive(ctx, 0, true)
	require.Error(t, err)

	qs.AssertNotCalled(t, "SetQuestionActive", mock.Anything, mock.Anything, mock.Anything)
}

func TestAdminService_SetQuestionActive_Success(t *testing.T) {
	qs := new(mockQuestionStore)
	svc := NewAdminService(qs)

	ctx := context.Background()

	expected := storage.QuestionRow{ID: 5, Text: "Q", IsActive: false}

	qs.On("SetQuestionActive", mock.Anything, int64(5), false).Return(expected, nil).Once()

	row, err := svc.SetQuestionActive(ctx, 5, false)
	require.NoError(t, err)
	require.Equal(t, expected, row)

	qs.AssertExpectations(t)
}
