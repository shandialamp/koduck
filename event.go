package koduck

import (
	"sync"
	"time"
)

const (
	ClientEventConnected = "connected"
)

const (
	ServerEventClientConnected    = "client_connected"
	ServerEventClientDisconnected = "client_disconnected"
)

type EventPayload interface {
	isEventPayload()
}

type ServerEventClientConnectedPayload struct {
	Time time.Time
	Conn *Conn
}

func (p *ServerEventClientConnectedPayload) isEventPayload() {}

type ServerEventClientDisconnectedPayload struct {
	Time    time.Time
	ConnKey string
}

func (p *ServerEventClientDisconnectedPayload) isEventPayload() {}

type ClientEventConnectedPayload struct{}

func (p *ClientEventConnectedPayload) isEventPayload() {}

type EventBus struct {
	subscribers map[string][]func(payload EventPayload) error
	mu          sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]func(payload EventPayload) error),
	}
}

func (bus *EventBus) Subscribe(eventName string, handler func(payload EventPayload) error) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.subscribers[eventName] = append(bus.subscribers[eventName], handler)
}

func (bus *EventBus) Publish(eventName string, payload EventPayload) {
	bus.mu.RLock()
	defer bus.mu.RUnlock()
	if handlers, ok := bus.subscribers[eventName]; ok {
		for _, handler := range handlers {
			handler(payload)
		}
	}
}

func (bus *EventBus) PublishAsync(eventName string, payload EventPayload) {
	bus.mu.RLock()
	defer bus.mu.RUnlock()
	if handlers, ok := bus.subscribers[eventName]; ok {
		for _, handler := range handlers {
			go handler(payload)
		}
	}
}
