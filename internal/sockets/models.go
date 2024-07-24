package sockets

import (
	"encoding/json"
	"quoridor/internal/game"
)

type EventType string

const (
	// IN
	EventTypeStartGame EventType = "start_game"
	EventTypeMakeMove  EventType = "make_move"
	EventTypePlaceWall EventType = "place_wall"
	EventTypeResign    EventType = "resign"
	EventTypeReconnect EventType = "reconnect"

	// OUT
	EventTypeGameState EventType = "game_state"
	EventTypeError     EventType = "error"
)

type WebsocketMessage struct {
	Type    EventType       `json:"event"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type MakeMovePayload struct {
	GameId   string        `json:"game_id"`
	Position game.Position `json:"position"`
}

type PlaceWallPayload struct {
	GameId string    `json:"game_id"`
	Wall   game.Wall `json:"wall"`
}

type ResignPayload struct {
	GameId string `json:"game_id"`
}

type ReconnectPayload struct {
	GameId string `json:"game_id"`
}

type ErrorMessagePayload struct {
	ErrorType string `json:"error_type"`
}
