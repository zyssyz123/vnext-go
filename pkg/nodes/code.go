package nodes

import (
	"dify-vnext-go/pkg/engine"
	"fmt"

	"github.com/dop251/goja"
)

type CodeNode struct {
	BaseNode
	Code string
}

func NewCodeNode(id string, config map[string]interface{}) *CodeNode {
	code, _ := config["code"].(string)
	return &CodeNode{
		BaseNode: NewBaseNode(id, "Code"),
		Code:     code,
	}
}

func (n *CodeNode) Execute(ctx *engine.NodeContext) (map[string]interface{}, error) {
	fmt.Printf("[%s] Executing Code...\n", n.ID())

	vm := goja.New()

	// Inject inputs into JS context
	for k, v := range ctx.Inputs {
		vm.Set(k, v)
	}

	// Execute code
	// We assume the code sets a variable 'output' or returns a value.
	// Dify usually expects a 'main' function.
	// For MVP, let's assume the code is just a script that returns an object.

	val, err := vm.RunString(n.Code)
	if err != nil {
		return nil, fmt.Errorf("code execution failed: %w", err)
	}

	// Convert result to map
	// If the result is an object, we return it as outputs.
	// If it's a primitive, we wrap it in "result".

	export := val.Export()
	outputs := make(map[string]interface{})

	if m, ok := export.(map[string]interface{}); ok {
		outputs = m
	} else {
		outputs["result"] = export
	}

	fmt.Printf("[%s] Code Result: %v\n", n.ID(), outputs)

	return outputs, nil
}
