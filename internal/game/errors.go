package game

import "errors"

var (
	ErrNotHost         = errors.New("not host")
	ErrBadPhase        = errors.New("bad phase")
	ErrNoPlayers       = errors.New("no players")
	ErrDeadlinePassed  = errors.New("deadline passed")
	ErrAlreadyAnswered = errors.New("already answered")
	ErrEmptyAnswer     = errors.New("empty answer")
	ErrInvalidOption   = errors.New("invalid option")
	ErrInvalidQuestion = errors.New("invalid question")
)
