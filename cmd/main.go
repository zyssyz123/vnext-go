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
	dslFile := flag.String("f", "examples/simple.yaml", "Path to DSL file")
	flag.Parse()

	// 1. Parse DSL
	fmt.Printf("Parsing DSL file: %s\n", *dslFile)
	workflow, err := dsl.Parse(*dslFile)
	if err != nil {
		log.Fatalf("Failed to parse DSL: %v", err)
	}
	fmt.Printf("Workflow loaded: %s (Version: %s)\n", workflow.Name, workflow.Version)

	// 2. Initialize Engine
	eng := engine.NewEngine(workflow)

	// 3. Register Nodes (Dynamic based on Workflow)
	for _, nodeDef := range workflow.Nodes {
		n := nodes.CreateNode(nodeDef)
		if n != nil {
			eng.RegisterNode(n)
		}
	}

	// 4. Run Workflow
	ctx := context.Background()
	initialInputs := map[string]interface{}{
		"query":  "Please search for the capital of France",
		"topics": []string{"Go Lang", "AI Agents", "Future Tech"},
		"topic":  "The Future of Quantum Computing",
	}

	fmt.Println("Starting workflow execution...")
	if err := eng.Run(ctx, initialInputs); err != nil {
		log.Fatalf("Workflow execution failed: %v", err)
	}
	fmt.Println("Workflow execution completed successfully.")
}
