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
	tools   map[string]Tool
	input   io.Reader
	output  io.Writer
	logger  *log.Logger
	ctx     context.Context
	cancel  context.CancelFunc
	mu      sync.Mutex
	running bool
}

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
	}
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
	s.mu.Lock()
	tool, exists := s.tools[req.Method]
	s.mu.Unlock()

	if !exists {
		s.sendErrorResponse(req.ID, CodeMethodNotFound, fmt.Sprintf("Method not found: %s", req.Method))
		return
	}

	result, err := tool.Handler(req.Params)
	if err != nil {
		s.sendErrorResponse(req.ID, CodeInternalError, err.Error())
		return
	}

	s.sendResponse(req.ID, result)
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
