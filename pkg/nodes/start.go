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
	fmt.Printf("[%s] Processing...\n", n.ID())

	// Populate memory with any inputs passed to the Start node (e.g. from YAML defaults)
	// Note: Engine.Run already populates memory with initialInputs (from CLI/API).
	// But if the YAML defines inputs for the Start node, they are resolved and passed here in ctx.Inputs.
	// We should merge them into memory so they are accessible via {{ memory.key }}.

	for k, v := range ctx.Inputs {
		ctx.Memory.Set(k, v)
	}

	// Return all memory as output
	return ctx.Memory.GetAll(), nil
}
