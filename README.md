# Sequential Thinking MCP Server

🧠 **Intelligent MCP Server** for step-by-step analysis and solving complex problems with support for **two operating modes**:
- **Stdio mode** for full compatibility with MCP clients (VS Code, Claude Desktop)
- **HTTP+SSE mode** for web interface and advanced debugging

## ⚡ Quick Start

### 1. Building the project
```bash
# Clone repository
git clone https://github.com/ad/sequentialthinking.git
cd sequentialthinking

# Local build using Go
go build -o sequentialthinking-server main.go

# Or using Make
make build-local

# Docker build
make build
```

### 2. Running the server

#### 📡 For MCP clients (VS Code, Claude Desktop)
```bash
./sequentialthinking-server --stdio
# Or using Make
make run-stdio
```

#### 🌐 For web development and debugging
```bash
./sequentialthinking-server
# Or with custom port
PORT=3000 ./sequentialthinking-server
# Or using Make
make run-local
```

#### 🐳 Using Docker
```bash
# Run in Docker
make run
# Or manually
docker run --rm -p 8080:8080 danielapatin/sequentialthinking:latest
```

### 3. Testing
```bash
# Automated testing
./test.sh

# Go unit tests
go test -v
# Or using Make
make test

# Web interface
open http://localhost:8080
```



## 🚀 Usage

### Integration with VS Code
Create a `.vscode/mcp.json` file in your project root:
```json
{
  "servers": {
    "sequentialthinking": {
      "type": "stdio", 
      "command": "/absolute/path/to/sequentialthinking-server",
      "args": ["--stdio"]
    }
  }
}
```

or use docker:

```json
{
  "servers": {
    "sequentialthinking": {
      "type": "stdio",
      "command": "docker",
      "args": ["run", "--rm", "-i", "danielapatin/sequentialthinking:latest", "--stdio"]
    },
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
      "args": ["--stdio"]
    }
  }
}
```

### Docker deployment
```bash
# Build image
docker build -t danielapatin/sequentialthinking:latest .
# Or using Make
make build

# Run container
docker run --rm -p 8080:8080 danielapatin/sequentialthinking:latest
# Or using Make
make run
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
echo '{"id": "1", "method": "tools/list"}' | ./sequentialthinking-server --stdio
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
- **Launch**: `./sequentialthinking-server --stdio`
- **Features**: Full compatibility with MCP specification 2025-03-26

### 🌐 HTTP+SSE Mode (Web Interface)
- **Purpose**: Debugging, testing, web integration
- **Protocol**: HTTP API + Server-Sent Events
- **Launch**: `./sequentialthinking-server` (default port 8080)
- **Endpoints**:
  - `GET /` - Web interface for testing
  - `POST /mcp` - MCP requests in JSON format  
  - `GET /events` - SSE real-time event stream

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
├── Makefile             # Build automation
├── test.sh              # Testing script
├── Dockerfile           # Docker configuration
├── README.md            # Documentation
├── sequentialthinking-server # Compiled binary
└── templates/
    └── index.html       # Web interface template
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
- **HTTP server port**: `PORT` environment variable (default 8080)
- **Logging**: all logs output to stderr
- **Operating mode**: determined by presence of `--stdio` flag

---

**✨ Project Status**: Active Development 
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
