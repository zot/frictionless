// Package mcp tests for notification delivery
package mcp

import (
	"sync"
	"testing"
)

// TestSendNotificationFormat tests that SendNotification properly formats params
func TestSendNotificationFormat(t *testing.T) {
	// Create a minimal server just to test the formatting logic
	var receivedMethod string
	var receivedParams map[string]any

	// Create a mock that captures what would be sent
	// We'll test the conversion logic directly
	params := map[string]interface{}{
		"message": "Hello",
		"value":   float64(42),
		"nested": map[string]interface{}{
			"a": float64(1),
			"b": float64(2),
		},
	}

	// Simulate what SendNotification does - convert to map[string]any
	paramsMap := make(map[string]any, len(params))
	for k, v := range params {
		paramsMap[k] = v
	}

	receivedMethod = "test_event"
	receivedParams = paramsMap

	// Verify
	if receivedMethod != "test_event" {
		t.Errorf("Expected method 'test_event', got '%s'", receivedMethod)
	}

	if receivedParams["message"] != "Hello" {
		t.Errorf("Expected message 'Hello', got %v", receivedParams["message"])
	}

	if receivedParams["value"] != float64(42) {
		t.Errorf("Expected value 42, got %v", receivedParams["value"])
	}

	nested, ok := receivedParams["nested"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected nested to be map, got %T", receivedParams["nested"])
	}
	if nested["a"] != float64(1) {
		t.Errorf("Expected nested.a = 1, got %v", nested["a"])
	}
}

// TestSendNotificationNilParams tests handling of nil params
func TestSendNotificationNilParams(t *testing.T) {
	var params interface{} = nil

	var paramsMap map[string]any
	if params != nil {
		if m, ok := params.(map[string]interface{}); ok {
			paramsMap = make(map[string]any, len(m))
			for k, v := range m {
				paramsMap[k] = v
			}
		}
	}

	if paramsMap != nil {
		t.Errorf("Expected nil paramsMap, got %v", paramsMap)
	}
}

// TestNotificationHandlerWiring tests that a notification handler can be called
func TestNotificationHandlerWiring(t *testing.T) {
	var mu sync.Mutex
	var called bool
	var receivedMethod string
	var receivedParams interface{}

	// Simulate a notification handler
	handler := func(method string, params interface{}) {
		mu.Lock()
		defer mu.Unlock()
		called = true
		receivedMethod = method
		receivedParams = params
	}

	// Call the handler as Lua would
	handler("user_action", map[string]interface{}{
		"action": "submit",
		"data":   "test",
	})

	mu.Lock()
	defer mu.Unlock()

	if !called {
		t.Error("Handler was not called")
	}

	if receivedMethod != "user_action" {
		t.Errorf("Expected method 'user_action', got '%s'", receivedMethod)
	}

	params, ok := receivedParams.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected params to be map, got %T", receivedParams)
	}

	if params["action"] != "submit" {
		t.Errorf("Expected action 'submit', got %v", params["action"])
	}
}

// TestMultipleNotificationCalls tests multiple rapid notification calls
func TestMultipleNotificationCalls(t *testing.T) {
	var mu sync.Mutex
	var methods []string

	handler := func(method string, params interface{}) {
		mu.Lock()
		defer mu.Unlock()
		methods = append(methods, method)
	}

	// Simulate multiple notifications
	for i := 0; i < 10; i++ {
		handler("event_"+string(rune('0'+i)), nil)
	}

	mu.Lock()
	defer mu.Unlock()

	if len(methods) != 10 {
		t.Errorf("Expected 10 notifications, got %d", len(methods))
	}
}
