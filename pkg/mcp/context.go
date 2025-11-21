package mcp

import (
	"context"
)

// MCPContext represents the MCP (Model Context Protocol) context
type MCPContext interface {
	// GetContext returns the underlying Go context
	GetContext() context.Context
	
	// GetParameter retrieves a parameter by name
	GetParameter(name string) (interface{}, bool)
	
	// SetParameter sets a parameter with the given name
	SetParameter(name string, value interface{})
	
	// GetParameters returns all parameters
	GetParameters() map[string]interface{}
	
	// Clone creates a copy of the context
	Clone() MCPContext
}

// DefaultMCPContext provides a default implementation of MCPContext
type DefaultMCPContext struct {
	ctx context.Context
	params map[string]interface{}
}

// NewDefaultMCPContext creates a new default MCP context
func NewDefaultMCPContext(ctx context.Context) *DefaultMCPContext {
	if ctx == nil {
		ctx = context.Background()
	}
	
	return &DefaultMCPContext{
		ctx:    ctx,
		params: make(map[string]interface{}),
	}
}

// GetContext returns the underlying Go context
func (d *DefaultMCPContext) GetContext() context.Context {
	return d.ctx
}

// GetParameter retrieves a parameter by name
func (d *DefaultMCPContext) GetParameter(name string) (interface{}, bool) {
	value, exists := d.params[name]
	return value, exists
}

// SetParameter sets a parameter with the given name
func (d *DefaultMCPContext) SetParameter(name string, value interface{}) {
	d.params[name] = value
}

// GetParameters returns all parameters
func (d *DefaultMCPContext) GetParameters() map[string]interface{} {
	return d.params
}

// Clone creates a copy of the context
func (d *DefaultMCPContext) Clone() MCPContext {
	clone := &DefaultMCPContext{
		ctx:    d.ctx,
		params: make(map[string]interface{}),
	}
	
	for k, v := range d.params {
		clone.params[k] = v
	}
	
	return clone
}