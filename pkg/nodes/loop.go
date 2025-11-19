package nodes

import (
	"encoding/json"
	"fmt"
	"sync"

	"dify-vnext-go/pkg/dsl"
	"dify-vnext-go/pkg/engine"
)

type LoopNode struct {
	BaseNode
	SubWorkflow  *dsl.WorkflowDefinition
	ParentEngine *engine.Engine // We need reference to parent engine to copy registry.
	// Actually we can't pass ParentEngine easily via constructor unless we change main.
	// Alternative: Pass it in Context? No, NodeContext has Memory but not Engine.
	// Let's assume we can get registry from somewhere or just require it to be passed.
	// Wait, NodeContext doesn't have Engine.
	// We might need to update NodeContext to include the Engine or Registry.
}

// We need a way to get the registry.
// Let's update NodeContext in pkg/engine/node.go to include the Engine or Registry.

func NewLoopNode(id string, config map[string]interface{}) *LoopNode {
	// Parse sub_workflow from config
	subWfMap, ok := config["sub_workflow"]
	if !ok {
		fmt.Printf("Error: sub_workflow missing in LoopNode config\n")
		return &LoopNode{BaseNode: NewBaseNode(id, "Loop")}
	}

	// Convert map to JSON then to struct (hacky but effective for dynamic map)
	jsonBytes, err := json.Marshal(subWfMap)
	if err != nil {
		fmt.Printf("Error marshaling sub_workflow: %v\n", err)
		return &LoopNode{BaseNode: NewBaseNode(id, "Loop")}
	}

	var subWf dsl.WorkflowDefinition
	if err := json.Unmarshal(jsonBytes, &subWf); err != nil {
		fmt.Printf("Error unmarshaling sub_workflow: %v\n", err)
		return &LoopNode{BaseNode: NewBaseNode(id, "Loop")}
	}

	return &LoopNode{
		BaseNode:    NewBaseNode(id, "Loop"),
		SubWorkflow: &subWf,
	}
}

func (n *LoopNode) Execute(ctx *engine.NodeContext) (map[string]interface{}, error) {
	// 1. Get Input List
	listInput, ok := ctx.Inputs["list"]
	if !ok {
		return nil, fmt.Errorf("missing input 'list'")
	}

	// Handle different list types (generic slice)
	var items []interface{}
	switch v := listInput.(type) {
	case []interface{}:
		items = v
	case []string:
		for _, s := range v {
			items = append(items, s)
		}
	default:
		return nil, fmt.Errorf("input 'list' must be an array, got %T", listInput)
	}

	fmt.Printf("[%s] Starting Loop over %d items...\n", n.ID(), len(items))

	// 2. Prepare Concurrency
	var wg sync.WaitGroup
	results := make([]interface{}, len(items))
	errCh := make(chan error, len(items))

	// 3. Iterate and Spawn Engines
	for i, item := range items {
		wg.Add(1)
		go func(index int, val interface{}) {
			defer wg.Done()

			// Create Sub-Engine
			subEngine := engine.NewEngine(n.SubWorkflow)

			// Register nodes for the sub-workflow
			for _, nodeDef := range n.SubWorkflow.Nodes {
				nodeInstance := CreateNode(nodeDef)
				if nodeInstance != nil {
					subEngine.RegisterNode(nodeInstance)
				}
			}

			// Also register nodes from parent engine if needed (e.g. shared tools?)
			// For now, let's assume sub-workflow is self-contained or uses standard nodes.
			// If we want to share stateful nodes, we might need to copy.
			// But CreateNode creates NEW instances, which is correct for Loop (isolation).

			// Inject Loop Item into Memory using Child Scope
			childMem := ctx.Memory.NewChild()
			childMem.Set("loop_item", val)

			// Set child memory scope to sub-engine
			subEngine.SetMemory(childMem)

			// Run Sub-Workflow
			// We pass empty inputs because we already populated the memory scope.
			if err := subEngine.Run(ctx.Ctx, nil); err != nil {
				errCh <- fmt.Errorf("iteration %d failed: %w", index, err)
				return
			}

			// Collect Results
			outputs := subEngine.GetOutputs()
			results[index] = outputs
		}(i, item)
	}

	wg.Wait()
	close(errCh)

	// Check errors
	if len(errCh) > 0 {
		return nil, <-errCh // Return first error
	}

	fmt.Printf("[%s] Loop completed.\n", n.ID())

	return map[string]interface{}{
		"results": results,
	}, nil
}
