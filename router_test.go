package koduck

import (
	"testing"
)

func TestRouterRegister(t *testing.T) {
	router := NewRouter()

	called := false
	handler := func(msg *Message) error {
		called = true
		return nil
	}

	router.Register(1000, handler)

	msg := &Message{Method: 1000}
	err := router.Handle(msg)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !called {
		t.Error("Handler was not called")
	}
}

func TestRouterHandleUnregistered(t *testing.T) {
	router := NewRouter()

	msg := &Message{Method: 9999}
	err := router.Handle(msg)

	if err == nil {
		t.Error("Expected error for unregistered method, got nil")
	}
}

func TestRegisterRouteWithGeneric(t *testing.T) {
	type Request struct {
		ID int `json:"id"`
	}

	router := NewRouter()
	var receivedID int

	RegisterRoute[Request](router, 2000, func(c *Conn, req *Request) error {
		receivedID = req.ID
		return nil
	})

	msg, _ := EncodeMessage(2000, &Request{ID: 42})
	err := router.Handle(msg)

	if err != nil {
		t.Fatalf("Failed to handle message: %v", err)
	}

	if receivedID != 42 {
		t.Errorf("Expected ID 42, got %d", receivedID)
	}
}

func TestRegisterRouteMultiple(t *testing.T) {
	router := NewRouter()
	count := 0

	RegisterRoute[string](router, 3000, func(c *Conn, s *string) error {
		count++
		return nil
	})

	msg, _ := EncodeMessage(3000, "test")
	router.Handle(msg)
	router.Handle(msg)

	if count != 2 {
		t.Errorf("Expected handler to be called twice, got %d times", count)
	}
}
