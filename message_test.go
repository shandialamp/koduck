package koduck

import (
	"encoding/json"
	"testing"
)

func TestEncodeMessage(t *testing.T) {
	type TestData struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	data := &TestData{Name: "Alice", Age: 30}
	msg, err := EncodeMessage(1000, data)

	if err != nil {
		t.Fatalf("EncodeMessage failed: %v", err)
	}

	if msg.Method != 1000 {
		t.Errorf("Expected method 1000, got %d", msg.Method)
	}

	if len(msg.Params) == 0 {
		t.Error("Expected params not to be empty")
	}
}

func TestDecodeMessage(t *testing.T) {
	type TestData struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	data := &TestData{Name: "Bob", Age: 25}
	msg, _ := EncodeMessage(2000, data)
	raw, _ := msg.Encode()

	decoded, err := DecodeMessage(raw)
	if err != nil {
		t.Fatalf("DecodeMessage failed: %v", err)
	}

	if decoded.Method != 2000 {
		t.Errorf("Expected method 2000, got %d", decoded.Method)
	}

	var result TestData
	if err := json.Unmarshal(decoded.Params, &result); err != nil {
		t.Fatalf("Failed to unmarshal params: %v", err)
	}

	if result.Name != "Bob" || result.Age != 25 {
		t.Errorf("Expected Bob, 25; got %s, %d", result.Name, result.Age)
	}
}

func TestMessageEncodeDecodeRoundTrip(t *testing.T) {
	msg := &Message{
		Method: 3000,
		Params: json.RawMessage(`{"value":123}`),
	}

	encoded, err := msg.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	decoded, err := DecodeMessage(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if decoded.Method != msg.Method {
		t.Errorf("Method mismatch: expected %d, got %d", msg.Method, decoded.Method)
	}

	if string(decoded.Params) != string(msg.Params) {
		t.Errorf("Params mismatch: expected %s, got %s", string(msg.Params), string(decoded.Params))
	}
}

func TestDecodeMessageInvalidJSON(t *testing.T) {
	invalidData := []byte("invalid json")
	_, err := DecodeMessage(invalidData)

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}
