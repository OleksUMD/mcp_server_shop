package mcp

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

type ToolFunc func(ctx context.Context, args map[string]any) (CallToolResult, error)

type Registry struct {
	tools    map[string]Tool
	Handlers map[string]ToolFunc
	logger   *logrus.Logger
}

func NewRegistry(logger *logrus.Logger) *Registry {
	return &Registry{
		tools:    make(map[string]Tool),
		Handlers: make(map[string]ToolFunc),
		logger:   logger,
	}
}

func (r *Registry) Register(tool Tool, handler ToolFunc) {
	r.tools[tool.Name] = tool
	r.Handlers[tool.Name] = handler
	r.logger.WithField("tool", tool.Name).Debug("Registered tool")
}

func (r *Registry) ListTools() []Tool {
	tools := make([]Tool, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

func (r *Registry) ExecuteTool(ctx context.Context, name string, args map[string]any) (CallToolResult, error) {
	handler, ok := r.Handlers[name]
	if !ok {
		return CallToolResult{}, fmt.Errorf("tool not found: %s", name)
	}
	return handler(ctx, args)
}
