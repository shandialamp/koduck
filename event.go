package koduck

import (
	"sync"
	"time"
)

const (
	ServerEventError              = 1
	ServerEventStarted            = 2
	ServerEventClientConnected    = 3
	ServerEventClientDisconnected = 4

	ClientEventError        = 101
	ClientEventConnected    = 102
	ClientEventDisconnected = 103
)

type EventPayload interface {
	isEventPayload()
}

type ServerEventErrorPayload struct {
	Err  error
	Data map[string]any
}

func (p *ServerEventErrorPayload) isEventPayload() {}

type ClientEventErrorPayload struct {
	Err  error
	Data map[string]any
}

func (p *ClientEventErrorPayload) isEventPayload() {}

type ServerEventClientConnectedPayload struct {
	Time time.Time
	Conn *Conn
}

func (p *ServerEventClientConnectedPayload) isEventPayload() {}

type ServerEventClientDisconnectedPayload struct {
	Time     time.Time
	ConnAddr string
}

func (p *ServerEventClientDisconnectedPayload) isEventPayload() {}

type ClientEventConnectedPayload struct {
	Conn       *Conn
	ServerAddr string // 服务端地址（包括端口）
	LocalAddr  string // 本地地址（包括端口）
	Time       time.Time
}

func (p *ClientEventConnectedPayload) isEventPayload() {}

type ClientEventDisconnectedPayload struct {
	ServerAddr string // 连接的服务端地址
	Reason     string // 断开原因
	Time       time.Time
}

func (p *ClientEventDisconnectedPayload) isEventPayload() {}

type EventBus struct {
	subscribers map[int][]func(payload EventPayload) error
	mu          sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[int][]func(payload EventPayload) error),
	}
}

func (bus *EventBus) Subscribe(eventName int, handler func(payload EventPayload) error) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.subscribers[eventName] = append(bus.subscribers[eventName], handler)
}

func (bus *EventBus) Publish(eventName int, payload EventPayload) {
	bus.mu.RLock()
	defer bus.mu.RUnlock()
	if handlers, ok := bus.subscribers[eventName]; ok {
		for _, handler := range handlers {
			handler(payload)
		}
	}
}

func (bus *EventBus) PublishAsync(eventName int, payload EventPayload) {
	bus.mu.RLock()
	defer bus.mu.RUnlock()
	if handlers, ok := bus.subscribers[eventName]; ok {
		for _, handler := range handlers {
			go handler(payload)
		}
	}
}
