package nodes

import (
	"dify-vnext-go/pkg/engine"
	"fmt"
	"strings"
)

type IfElseNode struct {
	BaseNode
	Conditions []Condition
}

type Condition struct {
	Variable string
	Operator string
	Value    string
}

func NewIfElseNode(id string, config map[string]interface{}) *IfElseNode {
	// Parse config for conditions
	// For MVP, we'll assume a simple "variable", "operator", "value" in config
	// or just a single condition for now.
	// Let's assume config has "variable", "operator", "value"

	v, _ := config["variable"].(string)
	op, _ := config["operator"].(string)
	val, _ := config["value"].(string)

	return &IfElseNode{
		BaseNode: NewBaseNode(id, "IfElse"),
		Conditions: []Condition{
			{Variable: v, Operator: op, Value: val},
		},
	}
}

func (n *IfElseNode) Execute(ctx *engine.NodeContext) (map[string]interface{}, error) {
	// Evaluate condition
	// For MVP, simple string comparison

	// Resolve variable from inputs
	// The DSL should map the variable to an input key, e.g. "input_1"
	// Or we just look at the first input?
	// Let's assume the DSL maps the value to check into an input named "input"

	inputVal, ok := ctx.Inputs["input"]
	if !ok {
		// Try to find it in the condition variable if it was passed as a template
		// But usually Inputs are resolved by Engine.
		// Let's assume the user mapped the variable to "input" in DSL.
		return nil, fmt.Errorf("missing input 'input' for IfElse node")
	}

	inputStr := fmt.Sprintf("%v", inputVal)
	cond := n.Conditions[0]

	result := false
	switch cond.Operator {
	case "equals":
		result = inputStr == cond.Value
	case "contains":
		result = strings.Contains(inputStr, cond.Value)
	default:
		// default to equals
		result = inputStr == cond.Value
	}

	fmt.Printf("[%s] Condition: '%s' %s '%s' ? %v\n", n.ID(), inputStr, cond.Operator, cond.Value, result)

	branchID := "false"
	if result {
		branchID = "true"
	}

	return map[string]interface{}{
		"result":     result,
		"_branch_id": branchID,
	}, nil
}
