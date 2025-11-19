package nodes

import (
	"dify-vnext-go/pkg/engine"
	"fmt"
)

type StartNode struct {
	BaseNode
}

func NewStartNode(id string) *StartNode {
	return &StartNode{
		BaseNode: NewBaseNode(id, "Start"),
	}
}

func (n *StartNode) Execute(ctx *engine.NodeContext) (map[string]interface{}, error) {
	// For MVP, StartNode exposes all global memory (initial inputs) as outputs
	fmt.Printf("[%s] Processing...\n", n.ID())
	return ctx.Memory.GetAll(), nil
}
