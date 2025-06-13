package main

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var version string

//go:embed templates/index.html
var indexHTML string

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

func colorize(color, text string) string {
	return color + text + ColorReset
}

// ThoughtData represents the data structure for a single thought
type ThoughtData struct {
	Thought           string `json:"thought"`
	ThoughtNumber     int    `json:"thoughtNumber"`
	TotalThoughts     int    `json:"totalThoughts"`
	NextThoughtNeeded bool   `json:"nextThoughtNeeded"`
	IsRevision        *bool  `json:"isRevision,omitempty"`
	RevisesThought    *int   `json:"revisesThought,omitempty"`
	BranchFromThought *int   `json:"branchFromThought,omitempty"`
	BranchID          string `json:"branchId,omitempty"`
	NeedsMoreThoughts *bool  `json:"needsMoreThoughts,omitempty"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Minimum     *int   `json:"minimum,omitempty"`
}

type MCPRequest struct {
	ID     int    `json:"id"`
	Method string `json:"method"`
	Params Params `json:"params"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Params struct {
	Name            string                 `json:"name,omitempty"`
	Arguments       map[string]interface{} `json:"arguments,omitempty"`
	ProtocolVersion string                 `json:"protocolVersion,omitempty"`
	Capabilities    map[string]interface{} `json:"capabilities,omitempty"`
	ClientInfo      ClientInfo             `json:"clientInfo,omitempty"`
}

type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      ClientInfo             `json:"clientInfo"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeResult struct {
	ProtocolVersion string       `json:"protocolVersion"`
	Capabilities    Capabilities `json:"capabilities"`
	ServerInfo      ServerInfo   `json:"serverInfo"`
	Instructions    string       `json:"instructions,omitempty"`
}

type Capabilities struct {
	Logging   map[string]interface{} `json:"logging,omitempty"`
	Tools     map[string]interface{} `json:"tools,omitempty"`
	Prompts   map[string]interface{} `json:"prompts,omitempty"`
	Resources map[string]interface{} `json:"resources,omitempty"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ToolListResult struct {
	Tools []Tool `json:"tools"`
}

type ToolCallResult struct {
	Content []Content `json:"content"`
	IsError *bool     `json:"isError,omitempty"`
}

type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// SequentialThinkingServer is the main server instance
type SequentialThinkingServer struct {
	thoughtHistory  []ThoughtData
	branches        map[string][]ThoughtData
	workspaceRoots  []string
	clientInfo      ClientInfo
	protocolVersion string
	clients         map[string]chan MCPMessage
	stdioMode       bool
}

type MCPMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func NewSequentialThinkingServer() *SequentialThinkingServer {
	return &SequentialThinkingServer{
		thoughtHistory:  make([]ThoughtData, 0),
		branches:        make(map[string][]ThoughtData),
		workspaceRoots:  make([]string, 0),
		protocolVersion: "2025-03-26",
		clients:         make(map[string]chan MCPMessage),
		stdioMode:       false,
	}
}

func (s *SequentialThinkingServer) SetStdioMode(stdio bool) {
	s.stdioMode = stdio
}

func (s *SequentialThinkingServer) SetClientInfo(clientInfo ClientInfo, protocolVersion string) {
	s.clientInfo = clientInfo
	s.protocolVersion = protocolVersion
	log.Printf("MCP Client: %s v%s, Protocol: %s", clientInfo.Name, clientInfo.Version, protocolVersion)
}

func (s *SequentialThinkingServer) SetWorkspaceRoots(roots []string) {
	s.workspaceRoots = roots
	log.Printf("Workspace roots: %v", roots)
}

func (s *SequentialThinkingServer) AddClient(clientID string) chan MCPMessage {
	ch := make(chan MCPMessage, 10)
	s.clients[clientID] = ch
	log.Printf("Client %s connected. Total clients: %d", clientID, len(s.clients))
	return ch
}

func (s *SequentialThinkingServer) RemoveClient(clientID string) {
	if ch, exists := s.clients[clientID]; exists {
		close(ch)
		delete(s.clients, clientID)
		log.Printf("Client %s disconnected. Total clients: %d", clientID, len(s.clients))
	}
}

func (s *SequentialThinkingServer) BroadcastMessage(msg MCPMessage) {
	for clientID, ch := range s.clients {
		select {
		case ch <- msg:
		default:
			log.Printf("Warning: Client %s channel full, removing client", clientID)
			s.RemoveClient(clientID)
		}
	}
}

func (s *SequentialThinkingServer) validateThoughtData(args map[string]interface{}) (*ThoughtData, error) {
	thought, ok := args["thought"].(string)
	if !ok || thought == "" {
		return nil, fmt.Errorf("invalid thought: must be a string")
	}

	thoughtNumberFloat, ok := args["thoughtNumber"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid thoughtNumber: must be a number")
	}
	thoughtNumber := int(thoughtNumberFloat)

	totalThoughtsFloat, ok := args["totalThoughts"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid totalThoughts: must be a number")
	}
	totalThoughts := int(totalThoughtsFloat)

	nextThoughtNeeded, ok := args["nextThoughtNeeded"].(bool)
	if !ok {
		return nil, fmt.Errorf("invalid nextThoughtNeeded: must be a boolean")
	}

	data := &ThoughtData{
		Thought:           thought,
		ThoughtNumber:     thoughtNumber,
		TotalThoughts:     totalThoughts,
		NextThoughtNeeded: nextThoughtNeeded,
	}

	// Optional fields
	if isRevision, ok := args["isRevision"].(bool); ok {
		data.IsRevision = &isRevision
	}

	if revisesThoughtFloat, ok := args["revisesThought"].(float64); ok {
		revisesThought := int(revisesThoughtFloat)
		data.RevisesThought = &revisesThought
	}

	if branchFromThoughtFloat, ok := args["branchFromThought"].(float64); ok {
		branchFromThought := int(branchFromThoughtFloat)
		data.BranchFromThought = &branchFromThought
	}

	if branchID, ok := args["branchId"].(string); ok {
		data.BranchID = branchID
	}

	if needsMoreThoughts, ok := args["needsMoreThoughts"].(bool); ok {
		data.NeedsMoreThoughts = &needsMoreThoughts
	}

	return data, nil
}

func (s *SequentialThinkingServer) formatThought(thoughtData *ThoughtData) string {
	var prefix, context string

	if thoughtData.IsRevision != nil && *thoughtData.IsRevision {
		prefix = colorize(ColorYellow, "🔄 Revision")
		if thoughtData.RevisesThought != nil {
			context = fmt.Sprintf(" (revising thought %d)", *thoughtData.RevisesThought)
		}
	} else if thoughtData.BranchFromThought != nil {
		prefix = colorize(ColorGreen, "🌿 Branch")
		context = fmt.Sprintf(" (from thought %d, ID: %s)", *thoughtData.BranchFromThought, thoughtData.BranchID)
	} else {
		prefix = colorize(ColorBlue, "💭 Thought")
		context = ""
	}

	header := fmt.Sprintf("%s %d/%d%s", prefix, thoughtData.ThoughtNumber, thoughtData.TotalThoughts, context)

	// Calculate border length (without ANSI codes)
	headerLen := len(fmt.Sprintf("💭 Thought %d/%d%s", thoughtData.ThoughtNumber, thoughtData.TotalThoughts, context))
	thoughtLen := len(thoughtData.Thought)
	borderLen := headerLen
	if thoughtLen > headerLen {
		borderLen = thoughtLen
	}
	borderLen += 4

	border := strings.Repeat("─", borderLen)

	return fmt.Sprintf(`
┌%s┐
│ %s │
├%s┤
│ %-*s │
└%s┘`, border, header, border, borderLen-2, thoughtData.Thought, border)
}

func (s *SequentialThinkingServer) processThought(args map[string]interface{}) ToolCallResult {
	validatedInput, err := s.validateThoughtData(args)
	if err != nil {
		isError := true
		errorData := map[string]interface{}{
			"error":  err.Error(),
			"status": "failed",
		}
		jsonData, _ := json.MarshalIndent(errorData, "", "  ")
		return ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: string(jsonData),
			}},
			IsError: &isError,
		}
	}

	// Adjust totalThoughts if necessary
	if validatedInput.ThoughtNumber > validatedInput.TotalThoughts {
		validatedInput.TotalThoughts = validatedInput.ThoughtNumber
	}

	// Add to history
	s.thoughtHistory = append(s.thoughtHistory, *validatedInput)

	// Add to branch if necessary
	if validatedInput.BranchFromThought != nil && validatedInput.BranchID != "" {
		if s.branches[validatedInput.BranchID] == nil {
			s.branches[validatedInput.BranchID] = make([]ThoughtData, 0)
		}
		s.branches[validatedInput.BranchID] = append(s.branches[validatedInput.BranchID], *validatedInput)
	}

	// Format and output thought
	formattedThought := s.formatThought(validatedInput)
	log.Print(formattedThought)

	// Send formatted thought to all connected clients (HTTP mode only)
	if !s.stdioMode {
		s.BroadcastMessage(MCPMessage{
			Type: "thought",
			Data: map[string]interface{}{
				"formatted": formattedThought,
				"raw":       validatedInput,
			},
		})
	}

	// Build response with context information
	branchKeys := make([]string, 0, len(s.branches))
	for k := range s.branches {
		branchKeys = append(branchKeys, k)
	}

	result := map[string]interface{}{
		"thoughtNumber":        validatedInput.ThoughtNumber,
		"totalThoughts":        validatedInput.TotalThoughts,
		"nextThoughtNeeded":    validatedInput.NextThoughtNeeded,
		"branches":             branchKeys,
		"thoughtHistoryLength": len(s.thoughtHistory),
		"workspaceContext": map[string]interface{}{
			"roots":           s.workspaceRoots,
			"clientInfo":      s.clientInfo,
			"protocolVersion": s.protocolVersion,
		},
	}

	jsonData, _ := json.MarshalIndent(result, "", "  ")
	return ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: string(jsonData),
		}},
	}
}

func getSequentialThinkingTool() Tool {
	return Tool{
		Name: "sequentialthinking",
		Description: `A detailed tool for dynamic and reflective problem-solving through thoughts.
This tool helps analyze problems through a flexible thinking process that can adapt and evolve.
Each thought can build on, question, or revise previous insights as understanding deepens.

When to use this tool:
- Breaking down complex problems into steps
- Planning and design with room for revision
- Analysis that might need course correction
- Problems where the full scope might not be clear initially
- Problems that require a multi-step solution
- Tasks that need to maintain context over multiple steps
- Situations where irrelevant information needs to be filtered out

Key features:
- You can adjust total_thoughts up or down as you progress
- You can question or revise previous thoughts
- You can add more thoughts even after reaching what seemed like the end
- You can express uncertainty and explore alternative approaches
- Not every thought needs to build linearly - you can branch or backtrack
- Generates a solution hypothesis
- Verifies the hypothesis based on the Chain of Thought steps
- Repeats the process until satisfied
- Provides a correct answer

Parameters explained:
- thought: Your current thinking step, which can include:
* Regular analytical steps
* Revisions of previous thoughts
* Questions about previous decisions
* Realizations about needing more analysis
* Changes in approach
* Hypothesis generation
* Hypothesis verification
- next_thought_needed: True if you need more thinking, even if at what seemed like the end
- thought_number: Current number in sequence (can go beyond initial total if needed)
- total_thoughts: Current estimate of thoughts needed (can be adjusted up/down)
- is_revision: A boolean indicating if this thought revises previous thinking
- revises_thought: If is_revision is true, which thought number is being reconsidered
- branch_from_thought: If branching, which thought number is the branching point
- branch_id: Identifier for the current branch (if any)
- needs_more_thoughts: If reaching end but realizing more thoughts needed

You should:
1. Start with an initial estimate of needed thoughts, but be ready to adjust
2. Feel free to question or revise previous thoughts
3. Don't hesitate to add more thoughts if needed, even at the "end"
4. Express uncertainty when present
5. Mark thoughts that revise previous thinking or branch into new paths
6. Ignore information that is irrelevant to the current step
7. Generate a solution hypothesis when appropriate
8. Verify the hypothesis based on the Chain of Thought steps
9. Repeat the process until satisfied with the solution
10. Provide a single, ideally correct answer as the final output
11. Only set next_thought_needed to false when truly done and a satisfactory answer is reached`,
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"thought": {
					Type:        "string",
					Description: "Your current thinking step",
				},
				"nextThoughtNeeded": {
					Type:        "boolean",
					Description: "Whether another thought step is needed",
				},
				"thoughtNumber": {
					Type:        "integer",
					Description: "Current thought number",
					Minimum:     intPtr(1),
				},
				"totalThoughts": {
					Type:        "integer",
					Description: "Estimated total thoughts needed",
					Minimum:     intPtr(1),
				},
				"isRevision": {
					Type:        "boolean",
					Description: "Whether this revises previous thinking",
				},
				"revisesThought": {
					Type:        "integer",
					Description: "Which thought is being reconsidered",
					Minimum:     intPtr(1),
				},
				"branchFromThought": {
					Type:        "integer",
					Description: "Branching point thought number",
					Minimum:     intPtr(1),
				},
				"branchId": {
					Type:        "string",
					Description: "Branch identifier",
				},
				"needsMoreThoughts": {
					Type:        "boolean",
					Description: "If more thoughts are needed",
				},
			},
			Required: []string{"thought", "nextThoughtNeeded", "thoughtNumber", "totalThoughts"},
		},
	}
}

func intPtr(i int) *int {
	return &i
}

func (s *SequentialThinkingServer) handleHTTPRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	response := s.handleRequest(req)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *SequentialThinkingServer) handleSSE(w http.ResponseWriter, r *http.Request) {
	// Setup SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// Generate unique client ID
	clientID := fmt.Sprintf("client_%d", time.Now().UnixNano())

	// Add client
	clientChan := s.AddClient(clientID)
	defer s.RemoveClient(clientID)

	// Send welcome message
	fmt.Fprintf(w, "data: %s\n\n", mustMarshal(MCPMessage{
		Type: "connected",
		Data: map[string]interface{}{
			"clientID": clientID,
			"message":  "Connected to Sequential Thinking MCP Server",
		},
	}))

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	// Listen for messages for client
	for {
		select {
		case msg, ok := <-clientChan:
			if !ok {
				return
			}

			fmt.Fprintf(w, "data: %s\n\n", mustMarshal(msg))
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}

		case <-r.Context().Done():
			return
		}
	}
}

func mustMarshal(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (s *SequentialThinkingServer) handleRequest(req MCPRequest) MCPResponse {
	switch req.Method {
	case "initialize":
		log.Printf("Received initialize request from client: %s (version: %s)\n", req.Params.ClientInfo.Name, req.Params.ClientInfo.Version)

		// Save client information
		s.SetClientInfo(req.Params.ClientInfo, req.Params.ProtocolVersion)

		// Extract workspace roots if available
		if roots, ok := req.Params.Capabilities["roots"]; ok {
			if rootsMap, ok := roots.(map[string]interface{}); ok {
				if rootsList, ok := rootsMap["roots"].([]interface{}); ok {
					workspaceRoots := make([]string, 0, len(rootsList))
					for _, root := range rootsList {
						if rootStr, ok := root.(string); ok {
							workspaceRoots = append(workspaceRoots, rootStr)
						}
					}
					s.SetWorkspaceRoots(workspaceRoots)
				}
			}
		}

		// Send initialization info via SSE (HTTP mode only)
		if !s.stdioMode {
			s.BroadcastMessage(MCPMessage{
				Type: "initialize",
				Data: map[string]interface{}{
					"client":  req.Params.ClientInfo,
					"version": req.Params.ProtocolVersion,
				},
			})
		}

		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: InitializeResult{
				ProtocolVersion: req.Params.ProtocolVersion,
				ServerInfo: ServerInfo{
					Name:    "sequentialthinking",
					Version: version,
				},
				Instructions: "Welcome to the Sequential Thinking MCP Server! Use the 'sequentialthinking' tool to process your thoughts step by step.",
				Capabilities: Capabilities{
					Logging: map[string]interface{}{},
					Tools: map[string]interface{}{
						"listChanged": true,
					},
				},
			},
		}
	case "notifications/initialized", "initialized":
		// This is a notification, don't return a response
		log.Printf("Received initialized notification - MCP handshake complete")
		if !s.stdioMode {
			s.BroadcastMessage(MCPMessage{
				Type: "initialized",
				Data: map[string]interface{}{
					"message": "MCP handshake complete",
				},
			})
		}
		return MCPResponse{} // Empty response for notifications
	case "tools/list":
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: ToolListResult{
				Tools: []Tool{getSequentialThinkingTool()},
			},
		}
	case "tools/call":
		if req.Params.Name == "sequentialthinking" {
			result := s.processThought(req.Params.Arguments)
			return MCPResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  result,
			}
		}
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32602,
				Message: fmt.Sprintf("Unknown tool: %s", req.Params.Name),
			},
		}
	default:
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32601,
				Message: fmt.Sprintf("Method not found: %s", req.Method),
			},
		}
	}
}

func main() {
	if version == "" {
		version = "1.0.0" // Default version if not set
	}

	log.SetOutput(os.Stderr)
	log.Printf("Starting Sequential Thinking MCP Server...")

	server := NewSequentialThinkingServer()

	// Check command line arguments for stdio mode
	useStdio := false
	for _, arg := range os.Args[1:] {
		if arg == "--stdio" {
			useStdio = true
			break
		}
	}

	if useStdio {
		log.Printf("Running in stdio mode")
		runStdioMode(server)
	} else {
		log.Printf("Running in HTTP+SSE mode")
		runHTTPMode(server)
	}
}

func runStdioMode(server *SequentialThinkingServer) {
	server.SetStdioMode(true)
	scanner := bufio.NewScanner(os.Stdin)

	requestCounter := 0
	for scanner.Scan() {
		requestCounter++
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		var req MCPRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			log.Printf("Error parsing request JSON: %v", err)

			var rawId struct {
				ID interface{} `json:"id"`
			}
			idStr := "null"
			if unmarshalIdErr := json.Unmarshal([]byte(line), &rawId); unmarshalIdErr == nil {
				switch v := rawId.ID.(type) {
				case string:
					idStr = v
				case float64:
					idStr = fmt.Sprintf("%.0f", v)
				case nil:
					idStr = "null"
				default:
					idStr = fmt.Sprintf("%v", v)
				}
			}

			errorResponse := MCPResponse{
				JSONRPC: "2.0",
				ID:      idStr,
				Error: &MCPError{
					Code:    -32700,
					Message: "Parse error: " + err.Error(),
				},
			}

			if responseJSON, marshalErr := json.Marshal(errorResponse); marshalErr == nil {
				fmt.Println(string(responseJSON))
				os.Stdout.Sync()
			}
			continue
		}

		response := server.handleRequest(req)

		// Skip response for notifications
		if req.Method == "initialized" || req.Method == "notifications/initialized" {
			continue
		}

		if responseJSON, err := json.Marshal(response); err == nil {
			fmt.Println(string(responseJSON))
			os.Stdout.Sync()
		} else {
			log.Printf("Error marshaling response: %v", err)
			internalErrorResponse := MCPResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &MCPError{
					Code:    -32603,
					Message: "Internal server error: failed to marshal response",
				},
			}
			if errorResponseJSON, _ := json.Marshal(internalErrorResponse); errorResponseJSON != nil {
				fmt.Println(string(errorResponseJSON))
				os.Stdout.Sync()
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Stdin scanner error: %v", err)
	}
	log.Printf("Sequential Thinking MCP Server shutting down.")
}

func runHTTPMode(server *SequentialThinkingServer) {
	// Setup HTTP routes
	http.HandleFunc("/mcp", server.handleHTTPRequest)
	http.HandleFunc("/events", server.handleSSE)
	http.HandleFunc("/", serveIndex)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting HTTP server on port %s", port)
	log.Printf("MCP endpoint: http://localhost:%s/mcp", port)
	log.Printf("SSE endpoint: http://localhost:%s/events", port)
	log.Printf("Web interface: http://localhost:%s/", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, indexHTML)
}
