# Sequential Thinking MCP Server

🧠 **Intelligent MCP Server** for step-by-step analysis and solving complex problems using the powerful [mcp-go](https://github.com/mark3labs/mcp-go) library.

Supports **multiple transport protocols**:
- **STDIO** - For MCP clients (VS Code, Claude Desktop)
- **SSE (Server-Sent Events)** - For real-time web applications  
- **StreamableHTTP** - For traditional web services and REST-like interactions

## ⚡ Quick Start

### 1. Building the project
```bash
# Clone repository
git clone https://github.com/ad/sequentialthinking.git
cd sequentialthinking

# Install dependencies
go mod tidy

# Local build using Go
go build -o sequentialthinking-server main.go

# Or using Make
make build-local

# Docker build
make build
```

### 2. Running the server

#### 📡 For MCP clients (VS Code, Claude Desktop) - STDIO mode
```bash
./sequentialthinking-server --stdio
# Or using Make
make run-stdio
```

#### 🌐 For real-time web applications - SSE mode
```bash
./sequentialthinking-server --sse
# Or with custom port
./sequentialthinking-server --sse --port 3000
```

#### 🔗 For traditional web services - HTTP mode
```bash
./sequentialthinking-server --http
# Or with custom port
./sequentialthinking-server --http -port 9090
```

#### 🐳 Using Docker
```bash
# Run in Docker (default STDIO mode)
make run
# Or manually
docker run --rm danielapatin/sequentialthinking:latest

# Run with HTTP mode on port 8080
docker run --rm -p 8080:8080 danielapatin/sequentialthinking:latest --http -port 8080
```

### 3. Testing
```bash
# Go unit tests
go test -v
# Or using Make
make test

# Test build
make build-local

# Test different transport modes
./sequentialthinking-server  -transport stdio   # Test STDIO mode
./sequentialthinking-server  -transport sse     # Test SSE mode on port 8080
./sequentialthinking-server  -transport http    # Test HTTP mode on port 8080
```

## 🚀 Usage

### Integration with VS Code
Add to `~/Library/Application Support/Code/User/settings.json`:

```json
{
  "mcp": {
    "servers": {
      "sequentialthinking": {
        "type": "stdio", 
        "command": "/absolute/path/to/sequentialthinking-server",
        "args": ["-transport", "stdio"]
      }
    }
  }
}
```

Or use Docker:

```json
{
  "mcp": {
    "servers": {
      "sequentialthinking": {
        "type": "stdio",
        "command": "docker",
        "args": ["run", "--rm", "-i", "danielapatin/sequentialthinking:latest", "-transport", "stdio"]
      }
    }
  }
}
```

### Integration with Claude Desktop
Add to `claude_desktop_config.json`:
```json
{
  "mcpServers": {
    "sequentialthinking": {
      "command": "/absolute/path/to/sequentialthinking-server",
      "args": ["-transport", "stdio"],
    }
  }
}
```

### Web Integration (SSE/HTTP modes)
For web applications, you can use SSE or HTTP transport:

```javascript
// SSE Client Example
const eventSource = new EventSource('http://localhost:8080/sse');
eventSource.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log('Received:', data);
};

// HTTP Client Example
fetch('http://localhost:8080/mcp', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
    },
    body: JSON.stringify({
        jsonrpc: '2.0',
        id: 1,
        method: 'tools/call',
        params: {
            name: 'sequentialthinking',
            arguments: {
                thought: 'Analyzing the problem step by step',
                thoughtNumber: 1,
                totalThoughts: 3,
                nextThoughtNeeded: true
            }
        }
    })
});
```

### Docker deployment
```bash
# Build image
docker build -t danielapatin/sequentialthinking:latest .
# Or using Make
make build

# Run container (STDIO mode)
docker run --rm danielapatin/sequentialthinking:latest -transport stdio

# Run container (HTTP mode with port mapping)
docker run --rm -p 8080:8080 danielapatin/sequentialthinking:latest -transport http -port 8083

# Run container (SSE mode with port mapping)
docker run --rm -p 8080:8080 danielapatin/sequentialthinking:latest -transport sse -port 8084
```

### Make commands
```bash
make help                 # Show all available commands
make build               # Build Docker image
make build-local         # Build local binary
make run                 # Run in Docker
make run-local           # Run locally (HTTP mode)
make run-stdio           # Run in stdio mode
make test                # Run tests in Docker
```

## 🔍 Testing and Debugging

### Automated testing
```bash
# Run all tests
./test.sh

# Run Go unit tests
go test -v
# Or using Make
make test

# Manual stdio mode testing
echo '{"id": "1", "method": "tools/list"}' | ./sequentialthinking-server -transport stdio
```

### HTTP API testing
```bash
# Start server in background
PORT=8083 ./sequentialthinking-server &

# List available tools
curl -X POST http://localhost:8083/mcp \
  -H "Content-Type: application/json" \
  -d '{"id": "1", "method": "tools/list"}'

# Call sequential thinking
curl -X POST http://localhost:8083/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "id": "2", 
    "method": "tools/call",
    "params": {
      "name": "sequentialthinking",
      "arguments": {
        "thought": "Analyzing this problem step by step",
        "thoughtNumber": 1,
        "totalThoughts": 3,
        "nextThoughtNeeded": true
      }
    }
  }'

# Connect to SSE stream
curl -N http://localhost:8083/events
```

### Web interface
Open `http://localhost:8080` for interactive testing with visual interface.

## 🧠 Sequential Thinking Tool

Provides a structured approach to solving complex problems through step-by-step thinking.

### Main parameters:
- **`thought`** *(string)*: Current thinking step  
- **`thoughtNumber`** *(integer)*: Current thought number (starting from 1)
- **`totalThoughts`** *(integer)*: Estimated total number of steps
- **`nextThoughtNeeded`** *(boolean)*: Whether the next step is required

### Additional parameters:
- **`isRevision`** *(boolean)*: Whether this thought revises a previous one
- **`revisesThought`** *(integer)*: Number of the thought being revised
- **`branchFromThought`** *(integer)*: Branching point for alternative approaches  
- **`branchId`** *(string)*: Branch identifier
- **`needsMoreThoughts`** *(boolean)*: Indicator of the need for additional steps

### Usage examples:

#### Basic sequential thinking:
```json
{
  "thought": "Starting analysis of the sorting algorithm",
  "thoughtNumber": 1,
  "totalThoughts": 5, 
  "nextThoughtNeeded": true
}
```

#### Revising previous solution:
```json
{
  "thought": "Reconsidering the data structure choice given performance requirements",
  "thoughtNumber": 3,
  "totalThoughts": 6,
  "nextThoughtNeeded": true,
  "isRevision": true,
  "revisesThought": 2
}
```

## 🔧 Operating Modes and Architecture

### 📡 Stdio Mode (MCP Compatibility)
- **Purpose**: Integration with MCP clients (VS Code, Claude Desktop)
- **Protocol**: JSON-RPC over stdin/stdout
- **Launch**: `./sequentialthinking-server -transport stdio`
- **Features**: Full compatibility with MCP specification 2025-03-26

### 🌐 SSE Mode
- **Purpose**: Debugging, testing, web integration
- **Protocol**: Server-Sent Events
- **Launch**: `./sequentialthinking-server -transport sse` (default port 8080)
- **Endpoints**:
  - `POST /sse` - SSE real-time event stream

### 🌐 HTTP Mode
- **Purpose**: Debugging, testing, web integration
- **Protocol**: HTTP API
- **Launch**: `./sequentialthinking-server -transport http` (default port 8080)
- **Endpoints**:
  - `POST /mcp` - MCP requests in JSON format

### Dual-mode architecture advantages:
✅ **Flexibility**: One server for different usage scenarios  
✅ **Debugging**: Web interface for testing and monitoring  
✅ **Compatibility**: Full MCP standard support  
✅ **Scalability**: HTTP mode supports multiple connections



## 🛠️ Technical Details

### System Requirements
- **Go**: version 1.24 or newer
- **OS**: Linux, macOS, Windows  
- **Dependencies**: Only Go standard library (no external packages)

### Project Structure
```
sequentialthinking/
├── main.go              # Main server code
├── main_test.go         # Unit tests  
├── go.mod               # Go module
├── go.sum               # Go dependencies
├── Makefile             # Build automation
├── test.sh              # Testing script
├── Dockerfile           # Docker configuration
└── README.md            # Documentation
```

### Building and Deployment
```bash
# Clone and build
git clone https://github.com/ad/sequentialthinking.git
cd sequentialthinking

# Local build
go build -o sequentialthinking-server main.go
# Or using Make
make build-local

# Docker build
make build

# Cross-platform builds
GOOS=linux GOARCH=amd64 go build -o sequentialthinking-linux main.go
GOOS=windows GOARCH=amd64 go build -o sequentialthinking.exe main.go
GOOS=darwin GOARCH=arm64 go build -o sequentialthinking-macos main.go
```

### Configuration
- **HTTP server port**: `-port 8080` variable (default 8080)
- **Operating mode**: determined by presence of `-transport stdio` flag
- **Logging**: all logs output to stderr

---

**📝 License**: MIT License

---

## 📚 Additional Resources

- 🌐 **[Model Context Protocol](https://modelcontextprotocol.io/)** - Official MCP documentation
- 🐙 **[GitHub Copilot Chat](https://docs.github.com/en/copilot/github-copilot-chat)** - Copilot Chat documentation

## 🤝 Support

If you have questions or issues:
1. Check the documentation in this repository
2. Review server logs in stderr
3. Try the web interface for debugging
4. Create an issue in the project repository
