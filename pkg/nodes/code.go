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
	// Dify Python injects inputs as a dictionary named 'input' (or similar, let's check docs/code if needed)
	// But for our examples, we used 'input.security', so we MUST inject 'input' as a map.
	// Also, we previously injected them as globals. Let's do BOTH for backward compatibility/flexibility.

	vm.Set("input", ctx.Inputs)

	for k, v := range ctx.Inputs {
		vm.Set(k, v)
	}

	// Determine code to run
	// 1. Use code from Config (n.Code)
	// 2. If empty, check if "code" is provided in Inputs (dynamic code)
	codeToRun := n.Code
	if codeToRun == "" {
		if val, ok := ctx.Inputs["code"]; ok {
			if s, ok := val.(string); ok {
				codeToRun = s
			}
		}
	}

	if codeToRun == "" {
		return nil, fmt.Errorf("no code provided for CodeNode %s", n.ID())
	}

	// Execute code
	// We assume the code sets a variable 'output' or returns a value.
	// For MVP, let's assume the code is just a script that returns an object.

	// print(codeToRun)

	val, err := vm.RunString(codeToRun)
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
