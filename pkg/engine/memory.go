package engine

import (
	"fmt"
	"strings"
	"sync"
)

// Memory interface defines the contract for memory management with scoping support
type Memory interface {
	Set(key string, value interface{})
	Get(key string) (interface{}, bool)
	GetAll() map[string]interface{}
	NewChild() Memory
	ResolveTemplate(template string) (interface{}, error)
}

// memoryScope implements the Memory interface with hierarchical scoping
type memoryScope struct {
	mu     sync.RWMutex
	data   map[string]interface{}
	parent Memory
}

// NewGlobalMemory creates a new root memory scope
func NewGlobalMemory() Memory {
	return &memoryScope{
		data: make(map[string]interface{}),
	}
}

// NewChild creates a new child memory scope
func (s *memoryScope) NewChild() Memory {
	return &memoryScope{
		data:   make(map[string]interface{}),
		parent: s,
	}
}

// Set stores a value in the current memory scope
func (s *memoryScope) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// Get retrieves a value from the current scope or parent scopes
func (s *memoryScope) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	val, ok := s.data[key]
	s.mu.RUnlock()

	if ok {
		return val, true
	}

	if s.parent != nil {
		return s.parent.Get(key)
	}

	return nil, false
}

// GetAll returns a flattened view of all data from root to current scope
func (s *memoryScope) GetAll() map[string]interface{} {
	var result map[string]interface{}
	if s.parent != nil {
		result = s.parent.GetAll()
	} else {
		result = make(map[string]interface{})
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	for k, v := range s.data {
		result[k] = v
	}
	return result
}

// ResolveTemplate resolves a template string like "{{ memory.key }}"
// Note: Node output resolution (e.g. "{{ node_id.key }}") is handled by the Engine,
// but we keep this method here to resolve memory references within templates.
func (s *memoryScope) ResolveTemplate(template string) (interface{}, error) {
	// Simple check for template syntax {{ ... }}
	if len(template) > 4 && template[:2] == "{{" && template[len(template)-2:] == "}}" {
		key := template[2 : len(template)-2]
		key = trimSpace(key)

		// Check for memory reference
		if len(key) > 7 && key[:7] == "memory." {
			memKey := key[7:]
			val, ok := s.Get(memKey)
			if !ok {
				return nil, fmt.Errorf("memory key not found: %s", memKey)
			}
			return val, nil
		}

		// If it's not a memory reference, we return the template as is,
		// letting the Engine handle other types of resolution (like node outputs).
		return template, nil
	}
	return template, nil
}

func trimSpace(s string) string {
	return strings.TrimSpace(s)
}
