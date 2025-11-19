package engine

import (
	"fmt"
	"sync"
)

// GlobalMemory is a thread-safe key-value store for workflow execution
type GlobalMemory struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

// NewGlobalMemory creates a new GlobalMemory instance
func NewGlobalMemory() *GlobalMemory {
	return &GlobalMemory{
		data: make(map[string]interface{}),
	}
}

// Set stores a value in memory
func (m *GlobalMemory) Set(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

// Get retrieves a value from memory
func (m *GlobalMemory) Get(key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.data[key]
	return val, ok
}

// GetAll returns a copy of all data
func (m *GlobalMemory) GetAll() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	copy := make(map[string]interface{}, len(m.data))
	for k, v := range m.data {
		copy[k] = v
	}
	return copy
}

// ResolveTemplate resolves a template string like "{{ node_id.output_key }}" or "{{ memory.key }}"
// This is a simplified implementation for the MVP
func (m *GlobalMemory) ResolveTemplate(template string, nodeOutputs map[string]map[string]interface{}) (interface{}, error) {
	// Simple check for template syntax {{ ... }}
	if len(template) > 4 && template[:2] == "{{" && template[len(template)-2:] == "}}" {
		key := template[2 : len(template)-2]
		key = trimSpace(key)

		// Check for memory reference
		if len(key) > 7 && key[:7] == "memory." {
			memKey := key[7:]
			val, ok := m.Get(memKey)
			if !ok {
				return nil, fmt.Errorf("memory key not found: %s", memKey)
			}
			return val, nil
		}

		// Check for node output reference (e.g., node_id.output_key)
		// This requires splitting by dot
		// For MVP, we assume simple format: node_id.key
		// In a real implementation, this would need a proper expression parser
		// Here we rely on the caller (Engine) to handle node outputs lookup if needed,
		// but for now let's assume we can access some global state or pass it in.
		// Actually, for this MVP, let's handle it by passing a map of all node outputs so far.

		// ... implementation continues in Engine ...
		return nil, fmt.Errorf("unsupported template format in memory resolver: %s", template)
	}
	return template, nil
}

func trimSpace(s string) string {
	// Simple trim implementation
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for start < end && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
