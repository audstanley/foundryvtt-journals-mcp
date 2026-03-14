package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

// JSON-RPC 2.0 types
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id"`
}

type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
)

// Server implements the MCP JSON-RPC 2.0 server
type Server struct {
	tools        map[string]Tool
	input        io.Reader
	output       io.Writer
	logger       *log.Logger
	ctx          context.Context
	cancel       context.CancelFunc
	mu           sync.Mutex
	running      bool
	capabilities *Capabilities
}

// Capabilities represents MCP server capabilities
type Capabilities struct {
	Tools     *ToolsCapability     `json:"tools"`
	Resources *ResourcesCapability `json:"resources"`
	Prompts   *PromptsCapability   `json:"prompts"`
}

// ToolsCapability represents tools capability structure
type ToolsCapability struct {
	ListChanged bool `json:"listChanged"`
}

// ResourcesCapability represents resources capability structure
type ResourcesCapability struct {
	ListChanged bool `json:"listChanged"`
}

// PromptsCapability represents prompts capability structure
type PromptsCapability struct{}

// Tool represents an MCP tool
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
	Handler     func(params json.RawMessage) (interface{}, error)
}

// NewServer creates a new MCP server
func NewServer(input io.Reader, output io.Writer) *Server {
	ctx, cancel := context.WithCancel(context.Background())

	return &Server{
		tools:  make(map[string]Tool),
		input:  input,
		output: output,
		logger: log.New(os.Stderr, "[MCP] ", log.LstdFlags),
		ctx:    ctx,
		cancel: cancel,
		capabilities: &Capabilities{
			Tools:     &ToolsCapability{ListChanged: false},
			Resources: &ResourcesCapability{ListChanged: false},
			Prompts:   &PromptsCapability{},
		},
	}
}

// SetLogger sets a custom logger for the server
func (s *Server) SetLogger(logger *log.Logger) {
	s.logger = logger
}

// RegisterTool registers a new tool
func (s *Server) RegisterTool(tool Tool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools[tool.Name] = tool
}

// Start starts the server
func (s *Server) Start() error {
	s.running = true

	reader := bufio.NewReader(s.input)
	for s.running {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF && !s.running {
				return nil
			}
			return fmt.Errorf("read error: %w", err)
		}

		var req Request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			s.sendErrorResponse(req.ID, CodeParseError, "Parse error")
			continue
		}

		s.handleRequest(req)
	}

	return nil
}

// Stop stops the server
func (s *Server) Stop() {
	s.mu.Lock()
	s.running = false
	s.mu.Unlock()
	s.cancel()
}

// handleRequest processes a JSON-RPC request
func (s *Server) handleRequest(req Request) {
	// Handle initialize method
	if req.Method == "initialize" {
		s.handleInitialize(req)
		return
	}

	// Handle ping method
	if req.Method == "ping" {
		s.handlePing(req)
		return
	}

	// Handle tools/list method
	if req.Method == "tools/list" {
		s.handleToolsList(req)
		return
	}

	// Handle tools/call method
	if req.Method == "tools/call" {
		s.handleToolsCall(req)
		return
	}

	// Handle resources/list method
	if req.Method == "resources/list" {
		s.handleResourcesList(req)
		return
	}

	// Handle resources/read method
	if req.Method == "resources/read" {
		s.handleResourcesRead(req)
		return
	}

	// Handle prompts/list method
	if req.Method == "prompts/list" {
		s.handlePromptsList(req)
		return
	}

	// Handle prompts/get method
	if req.Method == "prompts/get" {
		s.handlePromptsGet(req)
		return
	}

	s.sendErrorResponse(req.ID, CodeMethodNotFound, fmt.Sprintf("Method not found: %s", req.Method))
}

// handleInitialize handles the initialize request
func (s *Server) handleInitialize(req Request) {
	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"serverInfo": map[string]interface{}{
			"name":    "fvtt-journal-mcp",
			"version": "1.0.0",
		},
		"capabilities": s.capabilities,
	}
	s.sendResponse(req.ID, result)
}

// handlePing handles the ping request
func (s *Server) handlePing(req Request) {
	s.sendResponse(req.ID, map[string]interface{}{})
}

// handleToolsList handles the tools/list request
func (s *Server) handleToolsList(req Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	toolList := make([]map[string]interface{}, 0, len(s.tools))
	for _, tool := range s.tools {
		toolList = append(toolList, map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		})
	}

	s.sendResponse(req.ID, map[string]interface{}{
		"tools": toolList,
	})
}

// handleToolsCall handles the tools/call request
func (s *Server) handleToolsCall(req Request) {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments,omitempty"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendErrorResponse(req.ID, CodeInvalidParams, "Invalid params")
		return
	}

	s.mu.Lock()
	tool, exists := s.tools[params.Name]
	s.mu.Unlock()

	if !exists {
		s.sendErrorResponse(req.ID, CodeMethodNotFound, fmt.Sprintf("Tool not found: %s", params.Name))
		return
	}

	result, err := tool.Handler(params.Arguments)
	if err != nil {
		s.sendErrorResponse(req.ID, CodeInternalError, err.Error())
		return
	}

	content := []map[string]interface{}{
		{"type": "text", "text": fmt.Sprintf("%v", result)},
	}

	s.sendResponse(req.ID, map[string]interface{}{
		"content": content,
	})
}

// handleResourcesList handles the resources/list request
func (s *Server) handleResourcesList(req Request) {
	s.sendResponse(req.ID, map[string]interface{}{
		"resources": []interface{}{},
	})
}

// handleResourcesRead handles the resources/read request
func (s *Server) handleResourcesRead(req Request) {
	s.sendErrorResponse(req.ID, CodeMethodNotFound, "Resources read not implemented")
}

// handlePromptsList handles the prompts/list request
func (s *Server) handlePromptsList(req Request) {
	s.sendResponse(req.ID, map[string]interface{}{
		"prompts": []interface{}{},
	})
}

// handlePromptsGet handles the prompts/get request
func (s *Server) handlePromptsGet(req Request) {
	s.sendErrorResponse(req.ID, CodeMethodNotFound, "Prompts get not implemented")
}

// sendResponse sends a JSON-RPC response
func (s *Server) sendResponse(id interface{}, result interface{}) {
	resp := Response{
		JSONRPC: "2.0",
		Result:  result,
		ID:      id,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		s.logger.Printf("Error marshaling response: %v", err)
		return
	}

	fmt.Fprintf(s.output, "%s\n", string(data))
}

// sendErrorResponse sends a JSON-RPC error response
func (s *Server) sendErrorResponse(id interface{}, code int, message string) {
	resp := Response{
		JSONRPC: "2.0",
		Error: &Error{
			Code:    code,
			Message: message,
		},
		ID: id,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		s.logger.Printf("Error marshaling error response: %v", err)
		return
	}

	fmt.Fprintf(s.output, "%s\n", string(data))
}

// StdioListener creates a stdio listener
func StdioListener() net.Listener {
	return &stdioListener{}
}

type stdioListener struct{}

func (l *stdioListener) Accept() (net.Conn, error) {
	// For stdio, we just return a connection to stdin/stdout
	// This is simplified - in production you'd want proper handling
	conn, _ := net.Pipe()
	return conn, nil
}

func (l *stdioListener) Addr() net.Addr {
	return &stdioAddr{}
}

func (l *stdioListener) Close() error {
	return nil
}

type stdioAddr struct{}

func (a *stdioAddr) Network() string {
	return "stdio"
}

func (a *stdioAddr) String() string {
	return "stdio"
}
