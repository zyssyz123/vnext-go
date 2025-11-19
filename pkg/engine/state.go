package engine

import (
	"fmt"
	"sync"
)

// BlobRef represents a reference to a large binary object
// This avoids passing large data around in memory
type BlobRef struct {
	ID       string `json:"blob_id"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
	// In a real system, Data would not be here, or would be a stream/reader.
	// For MVP, we might store it here or in a separate BlobStore.
	// Let's keep it simple: Data is here but we encourage passing the BlobRef struct.
	Data []byte `json:"-"`
}

// Checkpointer defines the interface for saving and loading workflow state
type Checkpointer interface {
	// Save persists the state of a workflow thread
	Save(threadID string, state map[string]interface{}) error
	// Load retrieves the state of a workflow thread
	Load(threadID string) (map[string]interface{}, error)
}

// InMemoryCheckpointer is a simple in-memory implementation of Checkpointer
type InMemoryCheckpointer struct {
	mu    sync.RWMutex
	store map[string]map[string]interface{} // threadID -> state
}

// NewInMemoryCheckpointer creates a new in-memory checkpointer
func NewInMemoryCheckpointer() *InMemoryCheckpointer {
	return &InMemoryCheckpointer{
		store: make(map[string]map[string]interface{}),
	}
}

// Save persists the state
func (c *InMemoryCheckpointer) Save(threadID string, state map[string]interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Deep copy state to simulate persistence and avoid reference issues
	// For MVP, shallow copy of the top map is enough if values are immutable or we don't modify them later.
	// But memory map is mutable.
	// Let's do a simple copy.
	savedState := make(map[string]interface{})
	for k, v := range state {
		savedState[k] = v
	}

	c.store[threadID] = savedState
	fmt.Printf("[Checkpointer] Saved state for thread %s: %d items\n", threadID, len(savedState))
	return nil
}

// Load retrieves the state
func (c *InMemoryCheckpointer) Load(threadID string) (map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	state, ok := c.store[threadID]
	if !ok {
		return nil, fmt.Errorf("thread not found: %s", threadID)
	}

	// Return copy
	loadedState := make(map[string]interface{})
	for k, v := range state {
		loadedState[k] = v
	}

	return loadedState, nil
}
