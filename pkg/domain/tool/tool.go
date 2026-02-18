package tool

import (
	"context"
)

// Tool defines the interface that all BA tools must implement
type Tool interface {
	// Name returns the tool name
	Name() string

	// Description returns what the tool does
	Description() string

	// Schema returns JSON schema for input validation
	Schema() string

	// Execute runs the tool with given input
	Execute(ctx context.Context, input string) (string, error)
}

// ToolManager manages all available tools
type ToolManager struct {
	tools map[string]Tool
}

// NewToolManager creates a new tool manager
func NewToolManager() *ToolManager {
	return &ToolManager{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the manager
func (tm *ToolManager) Register(tool Tool) {
	tm.tools[tool.Name()] = tool
}

// Get retrieves a tool by name
func (tm *ToolManager) Get(name string) (Tool, bool) {
	tool, exists := tm.tools[name]
	return tool, exists
}

// List returns all registered tools
func (tm *ToolManager) List() []Tool {
	tools := make([]Tool, 0, len(tm.tools))
	for _, tool := range tm.tools {
		tools = append(tools, tool)
	}
	return tools
}

// Execute runs a tool by name
func (tm *ToolManager) Execute(ctx context.Context, name, input string) (string, error) {
	tool, exists := tm.Get(name)
	if !exists {
		return "", &ToolNotFoundError{Name: name}
	}

	return tool.Execute(ctx, input)
}

// ToolNotFoundError is returned when a tool doesn't exist
type ToolNotFoundError struct {
	Name string
}

func (e *ToolNotFoundError) Error() string {
	return "tool not found: " + e.Name
}

// ErrRequiresInput is returned when the agent needs user input
type ErrRequiresInput struct {
	Question string
}

func (e *ErrRequiresInput) Error() string {
	return "waiting for user input: " + e.Question
}
