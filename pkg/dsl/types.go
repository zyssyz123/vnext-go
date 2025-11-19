package dsl

// WorkflowDefinition represents the top-level structure of the DSL
type WorkflowDefinition struct {
	Name    string           `yaml:"name"`
	Version string           `yaml:"version"`
	Memory  MemoryDefinition `yaml:"memory"`
	Nodes   []NodeDefinition `yaml:"nodes"`
	Edges   []EdgeDefinition `yaml:"edges"`
}

// MemoryDefinition defines the schema for global memory
type MemoryDefinition struct {
	Schema map[string]string `yaml:"schema"`
}

// NodeDefinition defines a single node in the workflow
type NodeDefinition struct {
	ID      string                 `yaml:"id"`
	Type    string                 `yaml:"type"`
	Config  map[string]interface{} `yaml:"config"`
	Inputs  map[string]string      `yaml:"inputs"` // Key: InputName, Value: Template/Reference
	Outputs map[string]string      `yaml:"outputs"`
}

// EdgeDefinition defines a connection between nodes
type EdgeDefinition struct {
	Source       string `yaml:"source"`
	Target       string `yaml:"target"`
	SourceHandle string `yaml:"source_handle,omitempty"` // For conditional branching
}
