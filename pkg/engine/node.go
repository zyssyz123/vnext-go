package engine

import (
	"context"
)

// NodeContext provides context for node execution
type NodeContext struct {
	Ctx    context.Context
	Memory Memory
	Inputs map[string]interface{}
	NodeID string
	Engine *Engine // Reference to the executing engine
}

// Node is the interface that all workflow nodes must implement
type Node interface {
	ID() string
	Type() string
	Execute(ctx *NodeContext) (map[string]interface{}, error)
}
