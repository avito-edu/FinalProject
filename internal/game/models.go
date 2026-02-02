package game

type Phase string

const (
	PhaseLobby     Phase = "lobby"
	PhaseAnswering Phase = "answering"
	PhaseResults   Phase = "results"
)

type Player struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
