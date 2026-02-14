package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

type ToolHandler func(args map[string]interface{}) (*CallToolResult, error)

type Server struct {
	Name     string
	Version  string
	tools    map[string]Tool
	handlers map[string]ToolHandler
}

func NewServer(name, version string) *Server {
	return &Server{
		Name:     name,
		Version:  version,
		tools:    make(map[string]Tool),
		handlers: make(map[string]ToolHandler),
	}
}

func (s *Server) RegisterTool(tool Tool, handler ToolHandler) {
	s.tools[tool.Name] = tool
	s.handlers[tool.Name] = handler
}

func (s *Server) ServeStdio() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			s.sendError(nil, -32700, "Parse error")
			continue
		}

		switch req.Method {
		case "tools/list":
			s.handleListTools(req.ID)
		case "tools/call":
			s.handleCallTool(req)
		default:
			// Ignore other methods or implement capabilities negotiation
		}
	}
}

func (s *Server) handleListTools(id interface{}) {
	var tools []Tool
	for _, t := range s.tools {
		tools = append(tools, t)
	}

	s.sendResponse(id, ListToolsResult{Tools: tools})
}

func (s *Server) handleCallTool(req Request) {
	var params CallToolParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, -32602, "Invalid params")
		return
	}

	handler, ok := s.handlers[params.Name]
	if !ok {
		s.sendError(req.ID, -32601, "Tool not found")
		return
	}

	result, err := handler(params.Arguments)
	if err != nil {
		s.sendError(req.ID, -32000, err.Error())
		return
	}

	s.sendResponse(req.ID, result)
}

func (s *Server) sendResponse(id interface{}, result interface{}) {
	resp := Response{
		JsonRPC: "2.0",
		Result:  result,
		ID:      id,
	}
	bytes, _ := json.Marshal(resp)
	fmt.Println(string(bytes))
}

func (s *Server) sendError(id interface{}, code int, message string) {
	resp := Response{
		JsonRPC: "2.0",
		Error: &Error{
			Code:    code,
			Message: message,
		},
		ID: id,
	}
	bytes, _ := json.Marshal(resp)
	fmt.Println(string(bytes))
}
