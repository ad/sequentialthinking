package main

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestSequentialThinkingServer_ListTools(t *testing.T) {
	server := NewSequentialThinkingServer()

	tools, err := server.ListTools(context.Background())
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	if len(tools) != 1 {
		t.Fatalf("Expected 1 tool, got %d", len(tools))
	}

	tool := tools[0]
	if tool.Name != "sequentialthinking" {
		t.Errorf("Expected tool name 'sequentialthinking', got '%s'", tool.Name)
	}

	if tool.Description == "" {
		t.Error("Tool description should not be empty")
	}

	// Check required properties
	required := tool.InputSchema.Required
	expectedRequired := []string{"thought", "nextThoughtNeeded", "thoughtNumber", "totalThoughts"}

	if len(required) != len(expectedRequired) {
		t.Errorf("Expected %d required properties, got %d", len(expectedRequired), len(required))
	}

	for _, prop := range expectedRequired {
		found := false
		for _, req := range required {
			if req == prop {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Required property '%s' not found", prop)
		}
	}
}

func TestValidateThoughtRequest(t *testing.T) {
	server := NewSequentialThinkingServer()

	tests := []struct {
		name    string
		req     ThoughtRequest
		wantErr bool
	}{
		{
			name: "valid basic request",
			req: ThoughtRequest{
				Thought:           "This is a test thought",
				NextThoughtNeeded: true,
				ThoughtNumber:     1,
				TotalThoughts:     3,
			},
			wantErr: false,
		},
		{
			name: "empty thought",
			req: ThoughtRequest{
				Thought:           "",
				NextThoughtNeeded: true,
				ThoughtNumber:     1,
				TotalThoughts:     3,
			},
			wantErr: true,
		},
		{
			name: "invalid thought number",
			req: ThoughtRequest{
				Thought:           "Test",
				NextThoughtNeeded: true,
				ThoughtNumber:     0,
				TotalThoughts:     3,
			},
			wantErr: true,
		},
		{
			name: "invalid total thoughts",
			req: ThoughtRequest{
				Thought:           "Test",
				NextThoughtNeeded: true,
				ThoughtNumber:     1,
				TotalThoughts:     0,
			},
			wantErr: true,
		},
		{
			name: "thought number exceeds total without needs more",
			req: ThoughtRequest{
				Thought:           "Test",
				NextThoughtNeeded: true,
				ThoughtNumber:     5,
				TotalThoughts:     3,
				NeedsMoreThoughts: false,
			},
			wantErr: true,
		},
		{
			name: "thought number exceeds total with needs more",
			req: ThoughtRequest{
				Thought:           "Test",
				NextThoughtNeeded: true,
				ThoughtNumber:     5,
				TotalThoughts:     3,
				NeedsMoreThoughts: true,
			},
			wantErr: false,
		},
		{
			name: "revision without revises thought",
			req: ThoughtRequest{
				Thought:           "Test",
				NextThoughtNeeded: true,
				ThoughtNumber:     1,
				TotalThoughts:     3,
				IsRevision:        true,
			},
			wantErr: true,
		},
		{
			name: "valid revision",
			req: ThoughtRequest{
				Thought:           "Test",
				NextThoughtNeeded: true,
				ThoughtNumber:     2,
				TotalThoughts:     3,
				IsRevision:        true,
				RevisesThought:    1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := server.validateThoughtRequest(&tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateThoughtRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCallTool(t *testing.T) {
	server := NewSequentialThinkingServer()

	// Test valid tool call
	req := ThoughtRequest{
		Thought:           "This is my first thought about the problem",
		NextThoughtNeeded: true,
		ThoughtNumber:     1,
		TotalThoughts:     3,
	}

	// Convert to map format that mcp-go uses
	args := map[string]interface{}{
		"thought":           req.Thought,
		"nextThoughtNeeded": req.NextThoughtNeeded,
		"thoughtNumber":     float64(req.ThoughtNumber),
		"totalThoughts":     float64(req.TotalThoughts),
	}

	toolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "sequentialthinking",
			Arguments: args,
		},
	}

	result, err := server.CallTool(context.Background(), toolRequest)
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	if len(result.Content) == 0 {
		t.Fatal("Result content is empty")
	}

	// Test unknown tool
	unknownRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "unknown",
			Arguments: args,
		},
	}

	_, err = server.CallTool(context.Background(), unknownRequest)
	if err == nil {
		t.Error("Expected error for unknown tool")
	}
}

func TestFormatThoughtResponse(t *testing.T) {
	server := NewSequentialThinkingServer()

	tests := []struct {
		name     string
		req      ThoughtRequest
		contains []string
	}{
		{
			name: "basic thought",
			req: ThoughtRequest{
				Thought:           "This is a test thought",
				NextThoughtNeeded: true,
				ThoughtNumber:     1,
				TotalThoughts:     3,
			},
			contains: []string{"Thought 1/3", "This is a test thought", "Continuing to next thought"},
		},
		{
			name: "final thought",
			req: ThoughtRequest{
				Thought:           "This is the final thought",
				NextThoughtNeeded: false,
				ThoughtNumber:     3,
				TotalThoughts:     3,
			},
			contains: []string{"Thought 3/3", "This is the final thought", "Thinking process completed"},
		},
		{
			name: "revision thought",
			req: ThoughtRequest{
				Thought:           "This is a revision",
				NextThoughtNeeded: true,
				ThoughtNumber:     2,
				TotalThoughts:     3,
				IsRevision:        true,
				RevisesThought:    1,
			},
			contains: []string{"Thought 2/3", "Revision of Thought 1", "This is a revision"},
		},
		{
			name: "branched thought",
			req: ThoughtRequest{
				Thought:           "This is a branched thought",
				NextThoughtNeeded: true,
				ThoughtNumber:     2,
				TotalThoughts:     3,
				BranchID:          "alternative",
			},
			contains: []string{"Thought 2/3", "Branch: alternative", "This is a branched thought"},
		},
		{
			name: "needs more thoughts",
			req: ThoughtRequest{
				Thought:           "This thought needs more exploration",
				NextThoughtNeeded: false,
				ThoughtNumber:     3,
				TotalThoughts:     3,
				NeedsMoreThoughts: true,
			},
			contains: []string{"Thought 3/3", "Additional thoughts may be needed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := server.formatThoughtResponse(&tt.req, "test-session")

			for _, expected := range tt.contains {
				if !contains(response, expected) {
					t.Errorf("Response does not contain expected text '%s'. Response: %s", expected, response)
				}
			}
		})
	}
}

func TestBranchingLogic(t *testing.T) {
	server := NewSequentialThinkingServer()

	// Create a request with branching
	req := ThoughtRequest{
		Thought:           "This is a branched thought",
		NextThoughtNeeded: true,
		ThoughtNumber:     2,
		TotalThoughts:     3,
		BranchID:          "alternative",
		BranchFromThought: 1,
	}

	// Convert to map format that mcp-go uses
	args := map[string]interface{}{
		"thought":           req.Thought,
		"nextThoughtNeeded": req.NextThoughtNeeded,
		"thoughtNumber":     float64(req.ThoughtNumber),
		"totalThoughts":     float64(req.TotalThoughts),
		"branchId":          req.BranchID,
		"branchFromThought": float64(req.BranchFromThought),
	}

	toolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "sequentialthinking",
			Arguments: args,
		},
	}

	_, err := server.CallTool(context.Background(), toolRequest)
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}

	// Check that the branch was recorded
	sessionFound := false
	for _, history := range server.history {
		if len(history.Branches) > 0 {
			if branch, exists := history.Branches["alternative"]; exists {
				if len(branch) == 1 && branch[0] == 2 {
					sessionFound = true
					break
				}
			}
		}
	}

	if !sessionFound {
		t.Error("Branch was not properly recorded in history")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			findInString(s, substr))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
