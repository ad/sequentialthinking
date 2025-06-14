package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ThoughtRequest represents the input parameters for the sequential thinking tool
type ThoughtRequest struct {
	Thought           string `json:"thought"`
	NextThoughtNeeded bool   `json:"nextThoughtNeeded"`
	ThoughtNumber     int    `json:"thoughtNumber"`
	TotalThoughts     int    `json:"totalThoughts"`
	IsRevision        bool   `json:"isRevision,omitempty"`
	RevisesThought    int    `json:"revisesThought,omitempty"`
	BranchFromThought int    `json:"branchFromThought,omitempty"`
	BranchID          string `json:"branchId,omitempty"`
	NeedsMoreThoughts bool   `json:"needsMoreThoughts,omitempty"`
}

// ThoughtHistory stores the chain of thoughts
type ThoughtHistory struct {
	Thoughts  []ThoughtRequest `json:"thoughts"`
	Branches  map[string][]int `json:"branches,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
}

// SequentialThinkingServer implements the MCP server for sequential thinking
type SequentialThinkingServer struct {
	history map[string]*ThoughtHistory
}

// NewSequentialThinkingServer creates a new sequential thinking server
func NewSequentialThinkingServer() *SequentialThinkingServer {
	return &SequentialThinkingServer{
		history: make(map[string]*ThoughtHistory),
	}
}

// ListTools returns the available tools
func (s *SequentialThinkingServer) ListTools(ctx context.Context) ([]mcp.Tool, error) {
	return []mcp.Tool{
		{
			Name:        "sequentialthinking",
			Description: "A detailed tool for dynamic and reflective problem-solving through thoughts.\nThis tool helps analyze problems through a flexible thinking process that can adapt and evolve.\nEach thought can build on, question, or revise previous insights as understanding deepens.",
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"thought": map[string]interface{}{
						"type":        "string",
						"description": "Your current thinking step",
					},
					"nextThoughtNeeded": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether another thought step is needed",
					},
					"thoughtNumber": map[string]interface{}{
						"type":        "integer",
						"minimum":     1,
						"description": "Current thought number",
					},
					"totalThoughts": map[string]interface{}{
						"type":        "integer",
						"minimum":     1,
						"description": "Estimated total thoughts needed",
					},
					"isRevision": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether this revises previous thinking",
					},
					"revisesThought": map[string]interface{}{
						"type":        "integer",
						"minimum":     1,
						"description": "Which thought is being reconsidered",
					},
					"branchFromThought": map[string]interface{}{
						"type":        "integer",
						"minimum":     1,
						"description": "Branching point thought number",
					},
					"branchId": map[string]interface{}{
						"type":        "string",
						"description": "Branch identifier",
					},
					"needsMoreThoughts": map[string]interface{}{
						"type":        "boolean",
						"description": "If more thoughts are needed",
					},
				},
				Required: []string{"thought", "nextThoughtNeeded", "thoughtNumber", "totalThoughts"},
			},
		},
	}, nil
}

// CallTool handles tool execution
func (s *SequentialThinkingServer) CallTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if request.Params.Name != "sequentialthinking" {
		return nil, fmt.Errorf("unknown tool: %s", request.Params.Name)
	}

	// Parse arguments from the map format that mcp-go uses
	var req ThoughtRequest

	// Convert arguments map to JSON and then to our struct
	if args, ok := request.Params.Arguments.(map[string]interface{}); ok {
		if thought, exists := args["thought"]; exists {
			if thoughtStr, ok := thought.(string); ok {
				req.Thought = thoughtStr
			}
		}
		if nextThoughtNeeded, exists := args["nextThoughtNeeded"]; exists {
			if nextBool, ok := nextThoughtNeeded.(bool); ok {
				req.NextThoughtNeeded = nextBool
			}
		}
		if thoughtNumber, exists := args["thoughtNumber"]; exists {
			if thoughtNum, ok := thoughtNumber.(float64); ok {
				req.ThoughtNumber = int(thoughtNum)
			}
		}
		if totalThoughts, exists := args["totalThoughts"]; exists {
			if totalNum, ok := totalThoughts.(float64); ok {
				req.TotalThoughts = int(totalNum)
			}
		}
		if isRevision, exists := args["isRevision"]; exists {
			if revisionBool, ok := isRevision.(bool); ok {
				req.IsRevision = revisionBool
			}
		}
		if revisesThought, exists := args["revisesThought"]; exists {
			if revisesNum, ok := revisesThought.(float64); ok {
				req.RevisesThought = int(revisesNum)
			}
		}
		if branchFromThought, exists := args["branchFromThought"]; exists {
			if branchNum, ok := branchFromThought.(float64); ok {
				req.BranchFromThought = int(branchNum)
			}
		}
		if branchId, exists := args["branchId"]; exists {
			if branchStr, ok := branchId.(string); ok {
				req.BranchID = branchStr
			}
		}
		if needsMoreThoughts, exists := args["needsMoreThoughts"]; exists {
			if needsBool, ok := needsMoreThoughts.(bool); ok {
				req.NeedsMoreThoughts = needsBool
			}
		}
	} else {
		// Fallback: try to unmarshal as JSON (for testing)
		argsBytes, err := json.Marshal(request.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal arguments: %w", err)
		}
		if err := json.Unmarshal(argsBytes, &req); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
	}

	// Validate input
	if err := s.validateThoughtRequest(&req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Process the thought
	sessionID := fmt.Sprintf("session_%d", time.Now().Unix())
	if s.history[sessionID] == nil {
		s.history[sessionID] = &ThoughtHistory{
			Thoughts:  []ThoughtRequest{},
			Branches:  make(map[string][]int),
			CreatedAt: time.Now(),
		}
	}

	s.history[sessionID].Thoughts = append(s.history[sessionID].Thoughts, req)

	// Handle branching
	if req.BranchID != "" {
		if s.history[sessionID].Branches[req.BranchID] == nil {
			s.history[sessionID].Branches[req.BranchID] = []int{}
		}
		s.history[sessionID].Branches[req.BranchID] = append(
			s.history[sessionID].Branches[req.BranchID],
			req.ThoughtNumber,
		)
	}

	// Format response
	response := s.formatThoughtResponse(&req, sessionID)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: response,
			},
		},
	}, nil
}

// validateThoughtRequest validates the thought request parameters
func (s *SequentialThinkingServer) validateThoughtRequest(req *ThoughtRequest) error {
	if req.Thought == "" {
		return fmt.Errorf("thought cannot be empty")
	}
	if req.ThoughtNumber < 1 {
		return fmt.Errorf("thought number must be positive")
	}
	if req.TotalThoughts < 1 {
		return fmt.Errorf("total thoughts must be positive")
	}
	if req.ThoughtNumber > req.TotalThoughts && !req.NeedsMoreThoughts {
		return fmt.Errorf("thought number cannot exceed total thoughts unless more thoughts are needed")
	}
	if req.IsRevision && req.RevisesThought < 1 {
		return fmt.Errorf("revises thought must be specified for revisions")
	}
	return nil
}

// formatThoughtResponse formats the response for a thought
func (s *SequentialThinkingServer) formatThoughtResponse(req *ThoughtRequest, sessionID string) string {
	response := fmt.Sprintf("🤔 **Thought %d/%d**", req.ThoughtNumber, req.TotalThoughts)

	if req.IsRevision {
		response += fmt.Sprintf(" (Revision of Thought %d)", req.RevisesThought)
	}

	if req.BranchID != "" {
		response += fmt.Sprintf(" [Branch: %s]", req.BranchID)
	}

	response += fmt.Sprintf("\n\n%s", req.Thought)

	if req.NextThoughtNeeded {
		response += "\n\n*Continuing to next thought...*"
	} else {
		response += "\n\n✅ **Thinking process completed**"

		// Add summary of the thinking process
		if history := s.history[sessionID]; history != nil && len(history.Thoughts) > 1 {
			response += fmt.Sprintf("\n\n📊 **Summary**: Completed %d thoughts", len(history.Thoughts))

			if len(history.Branches) > 0 {
				response += fmt.Sprintf(" across %d branches", len(history.Branches))
			}
		}
	}

	if req.NeedsMoreThoughts {
		response += "\n\n🔄 **Note**: Additional thoughts may be needed to fully explore this problem."
	}

	return response
}

// ListResources returns the available resources (none for this server)
func (s *SequentialThinkingServer) ListResources(ctx context.Context) ([]mcp.Resource, error) {
	return []mcp.Resource{}, nil
}

// ReadResource reads a resource (not implemented for this server)
func (s *SequentialThinkingServer) ReadResource(ctx context.Context, request mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return nil, fmt.Errorf("no resources available")
}

// ListPrompts returns the available prompts (none for this server)
func (s *SequentialThinkingServer) ListPrompts(ctx context.Context) ([]mcp.Prompt, error) {
	return []mcp.Prompt{}, nil
}

// GetPrompt gets a prompt (not implemented for this server)
func (s *SequentialThinkingServer) GetPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return nil, fmt.Errorf("no prompts available")
}

func main() {
	var transport = flag.String("transport", "stdio", "Transport type: stdio, sse, or http")
	var port = flag.String("port", "8080", "Port for SSE/HTTP servers")
	flag.Parse()

	// Create server with proper configuration
	mcpServer := server.NewMCPServer(
		"sequentialthinking",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithLogging(),
	)

	// Add the sequential thinking tool
	mcpServer.AddTool(
		mcp.NewTool("sequentialthinking",
			mcp.WithDescription("A detailed tool for dynamic and reflective problem-solving through thoughts.\nThis tool helps analyze problems through a flexible thinking process that can adapt and evolve.\nEach thought can build on, question, or revise previous insights as understanding deepens."),
			mcp.WithString("thought",
				mcp.Description("Your current thinking step"),
				mcp.Required(),
			),
			mcp.WithBoolean("nextThoughtNeeded",
				mcp.Description("Whether another thought step is needed"),
				mcp.Required(),
			),
			mcp.WithNumber("thoughtNumber",
				mcp.Description("Current thought number"),
				mcp.Required(),
			),
			mcp.WithNumber("totalThoughts",
				mcp.Description("Estimated total thoughts needed"),
				mcp.Required(),
			),
			mcp.WithBoolean("isRevision",
				mcp.Description("Whether this revises previous thinking"),
			),
			mcp.WithNumber("revisesThought",
				mcp.Description("Which thought is being reconsidered"),
			),
			mcp.WithNumber("branchFromThought",
				mcp.Description("Branching point thought number"),
			),
			mcp.WithString("branchId",
				mcp.Description("Branch identifier"),
			),
			mcp.WithBoolean("needsMoreThoughts",
				mcp.Description("If more thoughts are needed"),
			),
		),
		handleSequentialThinking,
	)

	switch *transport {
	case "stdio":
		log.Println("Starting MCP server with STDIO transport...")
		if err := server.ServeStdio(mcpServer); err != nil {
			log.Fatal("STDIO server error:", err)
		}

	case "sse":
		log.Printf("Starting MCP server with SSE transport on port %s...", *port)
		sseServer := server.NewSSEServer(mcpServer)

		http.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
			sseServer.ServeHTTP(w, r)
		})

		if err := http.ListenAndServe(":"+*port, nil); err != nil {
			log.Fatal("SSE server error:", err)
		}

	case "http":
		log.Printf("Starting MCP server with streamable HTTP transport on port %s...", *port)
		httpServer := server.NewStreamableHTTPServer(mcpServer)

		log.Printf("HTTP server listening on :%s/mcp", *port)
		if err := httpServer.Start(":" + *port); err != nil {
			log.Fatal("HTTP server error:", err)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown transport: %s\n", *transport)
		fmt.Fprintf(os.Stderr, "Usage: %s [-transport stdio|sse|http] [-port PORT]\n", os.Args[0])
		os.Exit(1)
	}
}

// Global server instance for tool handling
var globalServer = NewSequentialThinkingServer()

// handleSequentialThinking handles the sequential thinking tool calls
func handleSequentialThinking(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return globalServer.CallTool(ctx, request)
}
