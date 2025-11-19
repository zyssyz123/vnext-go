package nodes

import (
	"dify-vnext-go/pkg/engine"
	"fmt"
)

type EndNode struct {
	BaseNode
}

func NewEndNode(id string) *EndNode {
	return &EndNode{
		BaseNode: NewBaseNode(id, "End"),
	}
}

func (n *EndNode) Execute(ctx *engine.NodeContext) (map[string]interface{}, error) {
	result, _ := ctx.Inputs["result"].(string)
	fmt.Printf("[%s] Workflow finished. Final Result: %s\n", n.ID(), result)

	// Store final result in memory
	ctx.Memory.Set("final_answer", result)

	return map[string]interface{}{
		"final_result": result,
	}, nil
}
