package events

import "log"

type EventHandler func(event *Event)

type EventService interface {
	Publish(event *Event)
	RegisterHandler(eventType EventType, handler EventHandler)
}

type EventServiceImpl struct {
	handlers map[EventType][]EventHandler
}

func NewEventService() *EventServiceImpl {
	return &EventServiceImpl{
		handlers: make(map[EventType][]EventHandler),
	}
}

func (es *EventServiceImpl) Publish(event *Event) {
	log.Printf("Publishing event: type=%v", event.Type)
	if handlers, found := es.handlers[event.Type]; found {
		for _, handler := range handlers {
			handler(event)
		}
	}
}

func (es *EventServiceImpl) RegisterHandler(eventType EventType, handler EventHandler) {
	log.Printf("Registering handler for event type: %v", eventType)
	es.handlers[eventType] = append(es.handlers[eventType], handler)
}
