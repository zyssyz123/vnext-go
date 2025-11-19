package nodes

import (
	"dify-vnext-go/pkg/dsl"
	"dify-vnext-go/pkg/engine"
	"fmt"
)

// CreateNode creates a node instance based on the definition
func CreateNode(def dsl.NodeDefinition) engine.Node {
	switch def.Type {
	case "Start":
		return NewStartNode(def.ID)
	case "End":
		return NewEndNode(def.ID)
	case "LLM":
		return NewLLMNode(def.ID, def.Config)
	case "IfElse":
		return NewIfElseNode(def.ID, def.Config)
	case "HttpRequest":
		return NewHttpRequestNode(def.ID, def.Config)
	case "Code":
		return NewCodeNode(def.ID, def.Config)
	case "Answer":
		return NewAnswerNode(def.ID, def.Config)
	case "Tool":
		return NewToolNode(def.ID, def.Config)
	case "Loop":
		return NewLoopNode(def.ID, def.Config)
	default:
		fmt.Printf("Unknown node type: %s\n", def.Type)
		return nil
	}
}
