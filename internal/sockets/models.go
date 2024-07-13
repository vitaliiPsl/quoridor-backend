package sockets

import "encoding/json"

type EventType string

const (
	// IN
	EventTypeSendMessage EventType = "send_message"

	// OUT
	EventTypeReceiveMessage EventType = "receive_message"
)

type WebsocketMessage struct {
	Type    EventType       `json:"event"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type SendMessagePayload struct {
	ReceiverId string `json:"receiver_id"`
	Message    string `json:"message"`
}

type ReceiveMessagePayload struct {
	SenderId string `json:"sender_id"`
	Message  string `json:"message"`
}
