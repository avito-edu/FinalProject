package ws

import "encoding/json"

type Envelope struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type JoinPayload struct {
	Name string `json:"name"`
}

type SubmitAnswerPayload struct {
	OptionID string `json:"optionId"`
}

type clientMsg struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
