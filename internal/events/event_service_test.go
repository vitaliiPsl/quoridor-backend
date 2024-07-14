package events

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockHandler struct {
	mock.Mock
}

func (m *MockHandler) Handle(event *Event) {
	m.Called(event)
}

func TestEventService_Publish(t *testing.T) {
	eventService := NewEventService()

	mockHandler := new(MockHandler)
	mockHandler.On("Handle", mock.Anything).Return()

	eventType := EventType("test_event")
	eventService.RegisterHandler(eventType, mockHandler.Handle)

	event := &Event{
		Type: eventType,
		Data: "test_data",
	}

	eventService.Publish(event)

	mockHandler.AssertCalled(t, "Handle", event)
}

func TestEventService_RegisterHandler(t *testing.T) {
	eventService := NewEventService()

	called := false
	handler := func(event *Event) {
		called = true
	}

	eventType := EventType("test_event")
	eventService.RegisterHandler(eventType, handler)

	event := &Event{
		Type: eventType,
		Data: "test_data",
	}

	eventService.Publish(event)

	assert.True(t, called)
}

func TestEventService_MultipleHandlers(t *testing.T) {
	eventService := NewEventService()

	called1 := false
	handler1 := func(event *Event) {
		called1 = true
	}

	called2 := false
	handler2 := func(event *Event) {
		called2 = true
	}

	eventType := EventType("test_event")
	eventService.RegisterHandler(eventType, handler1)
	eventService.RegisterHandler(eventType, handler2)

	event := &Event{
		Type: eventType,
		Data: "test_data",
	}

	eventService.Publish(event)

	assert.True(t, called1)
	assert.True(t, called2)
}
