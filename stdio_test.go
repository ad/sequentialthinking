package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestStdioModeActual тестирует реальное поведение stdio режима
func TestStdioModeActual(t *testing.T) {
	t.Run("ScannerBehaviorWithMultipleLines", func(t *testing.T) {
		// Создаем строку с несколькими JSON запросами
		requests := []string{
			`{"id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","clientInfo":{"name":"test","version":"1.0.0"}}}`,
			`{"id":2,"method":"tools/list"}`,
			`{"id":3,"method":"tools/call","params":{"name":"sequentialthinking","arguments":{"thought":"Test","thoughtNumber":1,"totalThoughts":1,"nextThoughtNeeded":false}}}`,
		}

		input := strings.Join(requests, "\n")
		reader := strings.NewReader(input)
		scanner := bufio.NewScanner(reader)

		server := NewSequentialThinkingServer()
		server.SetStdioMode(true)

		processedRequests := 0
		responses := make([]MCPResponse, 0)

		// Симулируем цикл из runStdioMode
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			processedRequests++

			if line == "" {
				continue
			}

			var req MCPRequest
			if err := json.Unmarshal([]byte(line), &req); err != nil {
				t.Errorf("Failed to parse request %d: %v", processedRequests, err)
				continue
			}

			response := server.handleRequest(req)

			// Пропускаем ответ для notifications
			if req.Method == "initialized" || req.Method == "notifications/initialized" {
				continue
			}

			responses = append(responses, response)
		}

		// Проверяем ошибки сканера
		if err := scanner.Err(); err != nil {
			t.Errorf("Scanner error: %v", err)
		}

		// Проверяем, что обработаны все запросы
		if processedRequests != len(requests) {
			t.Errorf("Expected to process %d requests, but processed %d", len(requests), processedRequests)
		}

		// Проверяем, что получили ожидаемое количество ответов
		expectedResponses := 3 // initialize, tools/list, tools/call
		if len(responses) != expectedResponses {
			t.Errorf("Expected %d responses, got %d", expectedResponses, len(responses))
		}

		// Проверяем, что все ответы без ошибок
		for i, response := range responses {
			if response.Error != nil {
				t.Errorf("Response %d has error: %v", i, response.Error)
			}
		}
	})

	t.Run("ScannerBehaviorWithEOF", func(t *testing.T) {
		// Тестируем поведение при EOF после первого запроса
		input := `{"id":1,"method":"tools/list"}`
		reader := strings.NewReader(input)
		scanner := bufio.NewScanner(reader)

		server := NewSequentialThinkingServer()
		iterations := 0

		for scanner.Scan() {
			iterations++
			line := strings.TrimSpace(scanner.Text())

			var req MCPRequest
			if err := json.Unmarshal([]byte(line), &req); err != nil {
				t.Errorf("Failed to parse request: %v", err)
				continue
			}

			response := server.handleRequest(req)
			if response.Error != nil {
				t.Errorf("Request failed: %v", response.Error)
			}
		}

		// После EOF scanner.Scan() должен вернуть false
		if iterations != 1 {
			t.Errorf("Expected exactly 1 iteration, got %d", iterations)
		}

		// Проверяем, что EOF - не ошибка
		if err := scanner.Err(); err != nil {
			t.Errorf("Unexpected scanner error: %v", err)
		}
	})

	t.Run("ScannerBehaviorWithEmptyLines", func(t *testing.T) {
		// Тестируем поведение с пустыми строками
		input := `{"id":1,"method":"tools/list"}

{"id":2,"method":"tools/list"}

`
		reader := strings.NewReader(input)
		scanner := bufio.NewScanner(reader)

		server := NewSequentialThinkingServer()
		processedRequests := 0
		totalIterations := 0

		for scanner.Scan() {
			totalIterations++
			line := strings.TrimSpace(scanner.Text())

			if line == "" {
				continue // Пропускаем пустые строки
			}

			var req MCPRequest
			if err := json.Unmarshal([]byte(line), &req); err != nil {
				t.Errorf("Failed to parse request: %v", err)
				continue
			}

			processedRequests++
			response := server.handleRequest(req)
			if response.Error != nil {
				t.Errorf("Request %d failed: %v", processedRequests, response.Error)
			}
		}

		if processedRequests != 2 {
			t.Errorf("Expected to process 2 requests, got %d", processedRequests)
		}

		if totalIterations < 2 {
			t.Errorf("Expected at least 2 iterations (including empty lines), got %d", totalIterations)
		}
	})
}

// TestStdioModeEdgeCases тестирует крайние случаи
func TestStdioModeEdgeCases(t *testing.T) {
	t.Run("InvalidJSON", func(t *testing.T) {
		input := `{"id":1,"method":"tools/list"}
invalid json
{"id":2,"method":"tools/list"}`

		reader := strings.NewReader(input)
		scanner := bufio.NewScanner(reader)

		server := NewSequentialThinkingServer()
		validRequests := 0
		totalLines := 0

		for scanner.Scan() {
			totalLines++
			line := strings.TrimSpace(scanner.Text())

			if line == "" {
				continue
			}

			var req MCPRequest
			if err := json.Unmarshal([]byte(line), &req); err != nil {
				// Ожидаем ошибку для invalid json, но продолжаем
				continue
			}

			validRequests++
			response := server.handleRequest(req)
			if response.Error != nil {
				t.Errorf("Valid request failed: %v", response.Error)
			}
		}

		if validRequests != 2 {
			t.Errorf("Expected 2 valid requests, got %d", validRequests)
		}

		if totalLines != 3 {
			t.Errorf("Expected 3 total lines, got %d", totalLines)
		}
	})

	t.Run("LargeInput", func(t *testing.T) {
		// Тестируем с большим количеством запросов
		var buffer bytes.Buffer
		expectedCount := 100

		for i := 1; i <= expectedCount; i++ {
			req := fmt.Sprintf(`{"id":%d,"method":"tools/list"}`, i)
			buffer.WriteString(req + "\n")
		}

		reader := strings.NewReader(buffer.String())
		scanner := bufio.NewScanner(reader)

		server := NewSequentialThinkingServer()
		processedCount := 0

		start := time.Now()
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())

			if line == "" {
				continue
			}

			var req MCPRequest
			if err := json.Unmarshal([]byte(line), &req); err != nil {
				t.Errorf("Failed to parse request: %v", err)
				continue
			}

			processedCount++
			response := server.handleRequest(req)
			if response.Error != nil {
				t.Errorf("Request %d failed: %v", processedCount, response.Error)
			}
		}
		duration := time.Since(start)

		if processedCount != expectedCount {
			t.Errorf("Expected to process %d requests, got %d", expectedCount, processedCount)
		}

		// Проверяем производительность
		if duration > time.Second {
			t.Errorf("Processing took too long: %v", duration)
		}

		t.Logf("Processed %d requests in %v", processedCount, duration)
	})
}

// BenchmarkStdioProcessing бенчмарк для производительности
func BenchmarkStdioProcessing(b *testing.B) {
	server := NewSequentialThinkingServer()
	req := MCPRequest{
		ID:     1,
		Method: "tools/list",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		response := server.handleRequest(req)
		if response.Error != nil {
			b.Errorf("Request failed: %v", response.Error)
		}
	}
}
