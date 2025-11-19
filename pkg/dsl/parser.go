package dsl

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Parse loads and parses a workflow DSL file
func Parse(filename string) (*WorkflowDefinition, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var workflow WorkflowDefinition
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Basic validation
	if len(workflow.Nodes) == 0 {
		return nil, fmt.Errorf("workflow must have at least one node")
	}

	return &workflow, nil
}
