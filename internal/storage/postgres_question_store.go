package storage

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ArtemMoroz51/FinalProject/internal/game"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNoQuestions = errors.New("no active questions")

type PostgresQuestionStore struct {
	db *pgxpool.Pool
}

func NewPostgresQuestionStore(db *pgxpool.Pool) *PostgresQuestionStore {
	return &PostgresQuestionStore{db: db}
}

func (s *PostgresQuestionStore) GetRandomActive(ctx context.Context) (game.Question, error) {
	var text string
	var optionsJSON []byte
	var correctID string

	err := s.db.QueryRow(ctx, `
		SELECT text, options, correct_id
		FROM questions
		WHERE is_active = true
		ORDER BY random()
		LIMIT 1
	`).Scan(&text, &optionsJSON, &correctID)
	if err != nil {
		return game.Question{}, ErrNoQuestions
	}

	var opts []game.Option
	if err := json.Unmarshal(optionsJSON, &opts); err != nil {
		return game.Question{}, err
	}

	return game.Question{
		Text:      text,
		Options:   opts,
		CorrectID: correctID,
	}, nil
}
func (s *PostgresQuestionStore) CreateQuestion(ctx context.Context, in CreateQuestionInput) (QuestionRow, error) {

	optsJSON, err := json.Marshal(in.Options)
	if err != nil {
		return QuestionRow{}, err
	}

	var row QuestionRow
	var createdAt time.Time

	err = s.db.QueryRow(ctx, `
		INSERT INTO questions (text, options, correct_id, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, text, options, correct_id, is_active, created_at
	`, in.Text, optsJSON, in.CorrectID, in.IsActive).Scan(
		&row.ID, &row.Text, &optsJSON, &row.CorrectID, &row.IsActive, &createdAt,
	)
	if err != nil {
		return QuestionRow{}, err
	}

	var opts []game.Option
	if err := json.Unmarshal(optsJSON, &opts); err != nil {
		return QuestionRow{}, err
	}
	row.Options = opts
	row.CreatedAt = createdAt.Format(time.RFC3339)

	return row, nil
}

func (s *PostgresQuestionStore) ListQuestions(ctx context.Context, includeInactive bool) ([]QuestionRow, error) {
	q := `
		SELECT id, text, options, correct_id, is_active, created_at
		FROM questions
	`
	if !includeInactive {
		q += ` WHERE is_active = true`
	}
	q += ` ORDER BY created_at DESC, id DESC`

	rows, err := s.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]QuestionRow, 0)
	for rows.Next() {
		var r QuestionRow
		var optsJSON []byte
		var createdAt time.Time
		if err := rows.Scan(&r.ID, &r.Text, &optsJSON, &r.CorrectID, &r.IsActive, &createdAt); err != nil {
			return nil, err
		}

		var opts []game.Option
		if err := json.Unmarshal(optsJSON, &opts); err != nil {
			return nil, err
		}
		r.Options = opts
		r.CreatedAt = createdAt.Format(time.RFC3339)

		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *PostgresQuestionStore) SetQuestionActive(ctx context.Context, id int64, active bool) (QuestionRow, error) {
	var r QuestionRow
	var optsJSON []byte
	var createdAt time.Time

	err := s.db.QueryRow(ctx, `
		UPDATE questions
		SET is_active = $2
		WHERE id = $1
		RETURNING id, text, options, correct_id, is_active, created_at
	`, id, active).Scan(&r.ID, &r.Text, &optsJSON, &r.CorrectID, &r.IsActive, &createdAt)
	if err != nil {
		return QuestionRow{}, err
	}

	var opts []game.Option
	if err := json.Unmarshal(optsJSON, &opts); err != nil {
		return QuestionRow{}, err
	}
	r.Options = opts
	r.CreatedAt = createdAt.Format(time.RFC3339)

	return r, nil
}
