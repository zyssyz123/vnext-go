package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"dify-vnext-go/pkg/dsl"
	"dify-vnext-go/pkg/engine"
	"dify-vnext-go/pkg/nodes"
)

func main() {
	workflowFile := flag.String("f", "examples/simple.yaml", "Path to workflow YAML file")
	flag.Parse()

	// 1. Load Workflow Definition
	wf, err := dsl.Parse(*workflowFile)
	if err != nil {
		log.Fatalf("Failed to parse workflow: %v", err)
	}

	fmt.Printf("Loaded workflow: %s\n", wf.Name)

	// 2. Initialize Engine
	eng := engine.NewEngine(wf)

	// Initialize Checkpointer
	cp := engine.NewInMemoryCheckpointer()
	eng.SetCheckpointer(cp)

	// 3. Register Nodes
	// In a real app, this would be dynamic or plugin-based
	// For MVP, we manually register known node types
	// We need to create instances based on the definition
	for _, nodeDef := range wf.Nodes {
		nodeInstance := nodes.CreateNode(nodeDef)
		if nodeInstance != nil {
			eng.RegisterNode(nodeInstance)
		} else {
			log.Printf("Warning: Unknown node type '%s' for node '%s'", nodeDef.Type, nodeDef.ID)
		}
	}

	// 4. Prepare Initial Inputs
	// For MVP, we can hardcode or parse from CLI args
	inputs := map[string]interface{}{
		"topic": "Go Lang", // For research.yaml
		"query": "Go Lang", // For simple.yaml
	}

	// 5. Run Workflow
	ctx := context.Background()
	if err := eng.Run(ctx, inputs); err != nil {
		log.Fatalf("Workflow execution failed: %v", err)
	}

	fmt.Println("Workflow execution completed successfully.")
}
