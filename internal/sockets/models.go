package sockets

import "encoding/json"

type EventType string

const (
	// IN
	EventTypeStartGame EventType = "start_game"

	// OUT
	EventTypeGameState EventType = "game_state"
)

type WebsocketMessage struct {
	Type    EventType       `json:"event"`
	Payload json.RawMessage `json:"payload,omitempty"`
}
