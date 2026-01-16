package storage

import (
	"context"

	"github.com/ArtemMoroz51/FinalProject/internal/game"
)

type QuestionRow struct {
	ID        int64         `json:"id"`
	Text      string        `json:"text"`
	Options   []game.Option `json:"options"`
	CorrectID string        `json:"correctId"`
	IsActive  bool          `json:"isActive"`
	CreatedAt string        `json:"createdAt"`
}

type CreateQuestionInput struct {
	Text      string        `json:"text"`
	Options   []game.Option `json:"options"`
	CorrectID string        `json:"correctId"`
	IsActive  bool          `json:"isActive"`
}

type QuestionStore interface {
	GetRandomActive(ctx context.Context) (game.Question, error)

	CreateQuestion(ctx context.Context, in CreateQuestionInput) (QuestionRow, error)
	ListQuestions(ctx context.Context, includeInactive bool) ([]QuestionRow, error)
	SetQuestionActive(ctx context.Context, id int64, active bool) (QuestionRow, error)
}
