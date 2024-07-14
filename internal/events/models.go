package events

type EventType string

const (
	EventTypeMatchFound EventType = "match_found"
)

type Event struct {
	Type EventType
	Data interface{}
}