package koduck

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestEventBusSubscribe(t *testing.T) {
	bus := NewEventBus()
	var called bool

	bus.Subscribe(100, func(p EventPayload) error {
		called = true
		return nil
	})

	bus.Publish(100, &ServerEventErrorPayload{})

	if !called {
		t.Error("Handler was not called")
	}
}

func TestEventBusMultipleSubscribers(t *testing.T) {
	bus := NewEventBus()
	var count int64

	for i := 0; i < 5; i++ {
		bus.Subscribe(200, func(p EventPayload) error {
			atomic.AddInt64(&count, 1)
			return nil
		})
	}

	bus.Publish(200, &ServerEventErrorPayload{})

	finalCount := atomic.LoadInt64(&count)
	if finalCount != 5 {
		t.Errorf("Expected 5 handlers called, got %d", finalCount)
	}
}

func TestEventBusMultipleEvents(t *testing.T) {
	bus := NewEventBus()
	var event1Called, event2Called bool

	bus.Subscribe(300, func(p EventPayload) error {
		event1Called = true
		return nil
	})

	bus.Subscribe(400, func(p EventPayload) error {
		event2Called = true
		return nil
	})

	bus.Publish(300, &ServerEventErrorPayload{})
	bus.Publish(400, &ServerEventErrorPayload{})

	if !event1Called {
		t.Error("Event 1 handler was not called")
	}
	if !event2Called {
		t.Error("Event 2 handler was not called")
	}
}

func TestEventBusUnregisteredEvent(t *testing.T) {
	bus := NewEventBus()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("EventBus panicked: %v", r)
		}
	}()

	bus.Publish(999, &ServerEventErrorPayload{})
}

func TestEventBusPublishAsync(t *testing.T) {
	bus := NewEventBus()
	var count int64
	var wg sync.WaitGroup

	wg.Add(1)
	bus.Subscribe(500, func(p EventPayload) error {
		defer wg.Done()
		atomic.AddInt64(&count, 1)
		return nil
	})

	bus.PublishAsync(500, &ServerEventErrorPayload{})

	wg.Wait()
	finalCount := atomic.LoadInt64(&count)
	if finalCount != 1 {
		t.Errorf("Expected handler to be called once, got %d times", finalCount)
	}
}

func TestEventBusConcurrentPublish(t *testing.T) {
	bus := NewEventBus()
	var count int64
	var wg sync.WaitGroup

	bus.Subscribe(600, func(p EventPayload) error {
		atomic.AddInt64(&count, 1)
		return nil
	})

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bus.Publish(600, &ServerEventErrorPayload{})
		}()
	}

	wg.Wait()

	finalCount := atomic.LoadInt64(&count)
	if finalCount != 100 {
		t.Errorf("Expected handler to be called 100 times, got %d", finalCount)
	}
}
