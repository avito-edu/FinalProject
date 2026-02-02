package service

import (
	"context"
	"errors"
	"strings"

	"github.com/ArtemMoroz51/FinalProject/internal/storage"
)

type AdminService interface {
	CreateQuestion(ctx context.Context, in storage.CreateQuestionInput) (storage.QuestionRow, error)
	ListQuestions(ctx context.Context, includeInactive bool) ([]storage.QuestionRow, error)
	SetQuestionActive(ctx context.Context, id int64, active bool) (storage.QuestionRow, error)
}

type adminService struct {
	qs storage.QuestionStore
}

func NewAdminService(qs storage.QuestionStore) AdminService {
	return &adminService{qs: qs}
}

func (a *adminService) CreateQuestion(ctx context.Context, in storage.CreateQuestionInput) (storage.QuestionRow, error) {
	in.Text = strings.TrimSpace(in.Text)
	in.CorrectID = strings.TrimSpace(in.CorrectID)
	if in.Text == "" || in.CorrectID == "" || len(in.Options) != 4 {
		return storage.QuestionRow{}, errors.New("invalid question payload")
	}
	return a.qs.CreateQuestion(ctx, in)
}

func (a *adminService) ListQuestions(ctx context.Context, includeInactive bool) ([]storage.QuestionRow, error) {
	return a.qs.ListQuestions(ctx, includeInactive)
}

func (a *adminService) SetQuestionActive(ctx context.Context, id int64, active bool) (storage.QuestionRow, error) {
	if id <= 0 {
		return storage.QuestionRow{}, errors.New("invalid id")
	}
	return a.qs.SetQuestionActive(ctx, id, active)
}
