package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestSequentialThinkingServer(t *testing.T) {
	server := NewSequentialThinkingServer()

	// Тест валидации данных
	t.Run("ValidateThoughtData", func(t *testing.T) {
		validArgs := map[string]interface{}{
			"thought":           "This is a test thought",
			"thoughtNumber":     float64(1),
			"totalThoughts":     float64(3),
			"nextThoughtNeeded": true,
		}

		data, err := server.validateThoughtData(validArgs)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if data.Thought != "This is a test thought" {
			t.Errorf("Expected thought 'This is a test thought', got '%s'", data.Thought)
		}

		if data.ThoughtNumber != 1 {
			t.Errorf("Expected thoughtNumber 1, got %d", data.ThoughtNumber)
		}
	})

	// Тест обработки мысли
	t.Run("ProcessThought", func(t *testing.T) {
		args := map[string]interface{}{
			"thought":           "Testing the thought process",
			"thoughtNumber":     float64(1),
			"totalThoughts":     float64(2),
			"nextThoughtNeeded": true,
		}

		result := server.processThought(args)

		if result.IsError != nil && *result.IsError {
			t.Errorf("Expected no error, got error in result")
		}

		if len(result.Content) == 0 {
			t.Errorf("Expected content in result")
		}

		// Проверяем, что содержимое является валидным JSON
		var jsonData map[string]interface{}
		err := json.Unmarshal([]byte(result.Content[0].Text), &jsonData)
		if err != nil {
			t.Errorf("Expected valid JSON, got error: %v", err)
		}
	})

	// Тест обработки MCP запроса
	t.Run("HandleRequest", func(t *testing.T) {
		req := MCPRequest{
			ID:     2,
			Method: "tools/list",
		}

		response := server.handleRequest(req)

		if response.Error != nil {
			t.Errorf("Expected no error, got %v", response.Error)
		}

		if response.ID != 2 {
			t.Errorf("Expected ID 'test-1', got '%s'", response.ID)
		}
	})
}

func TestMain(m *testing.M) {
	// Настройка для тестов
	code := m.Run()
	os.Exit(code)
}
