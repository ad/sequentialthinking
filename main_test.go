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

// Тест для проверки обработки множественных запросов
func TestMultipleRequestProcessing(t *testing.T) {
	server := NewSequentialThinkingServer()

	t.Run("MultipleSequentialRequests", func(t *testing.T) {
		requests := []MCPRequest{
			{
				ID:     1,
				Method: "initialize",
				Params: Params{
					ProtocolVersion: "2025-03-26",
					ClientInfo: ClientInfo{
						Name:    "test-client",
						Version: "1.0.0",
					},
				},
			},
			{
				ID:     2,
				Method: "tools/list",
			},
			{
				ID:     3,
				Method: "tools/call",
				Params: Params{
					Name: "sequentialthinking",
					Arguments: map[string]interface{}{
						"thought":           "First thought",
						"thoughtNumber":     float64(1),
						"totalThoughts":     float64(3),
						"nextThoughtNeeded": true,
					},
				},
			},
			{
				ID:     4,
				Method: "tools/call",
				Params: Params{
					Name: "sequentialthinking",
					Arguments: map[string]interface{}{
						"thought":           "Second thought",
						"thoughtNumber":     float64(2),
						"totalThoughts":     float64(3),
						"nextThoughtNeeded": true,
					},
				},
			},
			{
				ID:     5,
				Method: "tools/call",
				Params: Params{
					Name: "sequentialthinking",
					Arguments: map[string]interface{}{
						"thought":           "Final thought",
						"thoughtNumber":     float64(3),
						"totalThoughts":     float64(3),
						"nextThoughtNeeded": false,
					},
				},
			},
		}

		// Обрабатываем все запросы последовательно
		responses := make([]MCPResponse, len(requests))
		for i, req := range requests {
			responses[i] = server.handleRequest(req)

			// Проверяем, что ответ корректный
			if responses[i].Error != nil {
				t.Errorf("Request %d failed with error: %v", i+1, responses[i].Error)
			}

			if responses[i].ID != req.ID {
				t.Errorf("Request %d: expected ID %v, got %v", i+1, req.ID, responses[i].ID)
			}
		}

		// Проверяем, что история мыслей накапливается
		if len(server.thoughtHistory) != 3 {
			t.Errorf("Expected 3 thoughts in history, got %d", len(server.thoughtHistory))
		}

		// Проверяем последовательность мыслей
		for i, thought := range server.thoughtHistory {
			expectedThoughtNumber := i + 1
			if thought.ThoughtNumber != expectedThoughtNumber {
				t.Errorf("Thought %d: expected thoughtNumber %d, got %d",
					i, expectedThoughtNumber, thought.ThoughtNumber)
			}
		}
	})

	t.Run("StdioModeSimulation", func(t *testing.T) {
		// Тестируем логику обработки запросов напрямую
		// вместо перенаправления stdin/stdout

		server := NewSequentialThinkingServer()
		server.SetStdioMode(true)

		// Готовим тестовые запросы
		testRequests := []MCPRequest{
			{
				ID:     1,
				Method: "initialize",
				Params: Params{
					ProtocolVersion: "2025-03-26",
					ClientInfo: ClientInfo{
						Name:    "test",
						Version: "1.0.0",
					},
				},
			},
			{
				ID:     2,
				Method: "tools/list",
			},
			{
				ID:     3,
				Method: "tools/call",
				Params: Params{
					Name: "sequentialthinking",
					Arguments: map[string]interface{}{
						"thought":           "Test thought",
						"thoughtNumber":     float64(1),
						"totalThoughts":     float64(1),
						"nextThoughtNeeded": false,
					},
				},
			},
		}

		responses := make([]MCPResponse, 0)

		// Обрабатываем запросы последовательно
		for _, req := range testRequests {
			response := server.handleRequest(req)
			
			// Пропускаем пустые ответы для уведомлений
			if req.Method != "initialized" && req.Method != "notifications/initialized" {
				responses = append(responses, response)
			}
		}

		// Проверяем, что получили ожидаемое количество ответов
		expectedResponses := 3 // initialize, tools/list, tools/call
		if len(responses) != expectedResponses {
			t.Errorf("Expected %d responses, got %d", expectedResponses, len(responses))
		}

		// Проверяем, что все ответы корректные
		for i, response := range responses {
			if response.Error != nil {
				t.Errorf("Response %d has error: %v", i, response.Error)
			}
		}
	})
}

// Тест для проверки обработки ошибок
func TestErrorHandling(t *testing.T) {
	server := NewSequentialThinkingServer()

	t.Run("InvalidMethod", func(t *testing.T) {
		req := MCPRequest{
			ID:     1,
			Method: "invalid/method",
		}

		response := server.handleRequest(req)

		if response.Error == nil {
			t.Error("Expected error for invalid method")
		}

		if response.Error.Code != -32601 {
			t.Errorf("Expected error code -32601, got %d", response.Error.Code)
		}
	})

	t.Run("InvalidTool", func(t *testing.T) {
		req := MCPRequest{
			ID:     2,
			Method: "tools/call",
			Params: Params{
				Name: "nonexistent",
			},
		}

		response := server.handleRequest(req)

		if response.Error == nil {
			t.Error("Expected error for invalid tool")
		}

		if response.Error.Code != -32602 {
			t.Errorf("Expected error code -32602, got %d", response.Error.Code)
		}
	})

	t.Run("InvalidThoughtData", func(t *testing.T) {
		req := MCPRequest{
			ID:     3,
			Method: "tools/call",
			Params: Params{
				Name: "sequentialthinking",
				Arguments: map[string]interface{}{
					"thought": "", // Пустая мысль должна вызвать ошибку
				},
			},
		}

		response := server.handleRequest(req)

		// Проверяем, что результат содержит ошибку
		if result, ok := response.Result.(ToolCallResult); ok {
			if result.IsError == nil || !*result.IsError {
				t.Error("Expected error result for invalid thought data")
			}
		} else {
			t.Error("Expected ToolCallResult in response")
		}
	})
}

// Бенчмарк для проверки производительности
func BenchmarkRequestHandling(b *testing.B) {
	server := NewSequentialThinkingServer()

	req := MCPRequest{
		ID:     1,
		Method: "tools/call",
		Params: Params{
			Name: "sequentialthinking",
			Arguments: map[string]interface{}{
				"thought":           "Benchmark thought",
				"thoughtNumber":     float64(1),
				"totalThoughts":     float64(1),
				"nextThoughtNeeded": false,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server.handleRequest(req)
	}
}
