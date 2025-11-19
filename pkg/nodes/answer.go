package nodes

import (
	"dify-vnext-go/pkg/engine"
	"fmt"
	"time"
)

type AnswerNode struct {
	BaseNode
}

func NewAnswerNode(id string, config map[string]interface{}) *AnswerNode {
	return &AnswerNode{
		BaseNode: NewBaseNode(id, "Answer"),
	}
}

func (n *AnswerNode) Execute(ctx *engine.NodeContext) (map[string]interface{}, error) {
	answer, _ := ctx.Inputs["answer"].(string)

	// Simulate streaming
	fmt.Printf("[%s] Streaming Answer: ", n.ID())
	for _, char := range answer {
		fmt.Printf("%c", char)
		time.Sleep(10 * time.Millisecond)
	}
	fmt.Println()

	return map[string]interface{}{
		"answer": answer,
	}, nil
}
