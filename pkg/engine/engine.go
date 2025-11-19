package engine

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"dify-vnext-go/pkg/dsl"
)

// Engine is the main runtime engine
type Engine struct {
	workflow *dsl.WorkflowDefinition
	memory   *GlobalMemory
	nodes    map[string]Node
	outputs  map[string]map[string]interface{} // node_id -> outputs
	mu       sync.RWMutex
}

// NewEngine creates a new engine instance
func NewEngine(wf *dsl.WorkflowDefinition) *Engine {
	return &Engine{
		workflow: wf,
		memory:   NewGlobalMemory(),
		nodes:    make(map[string]Node),
		outputs:  make(map[string]map[string]interface{}),
	}
}

// GetOutputs returns the outputs of all nodes
func (e *Engine) GetOutputs() map[string]map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()
	// Return a copy to be safe? For MVP just return the map, caller shouldn't modify.
	return e.outputs
}

// RegisterNode registers a node implementation
func (e *Engine) RegisterNode(n Node) {
	e.nodes[n.ID()] = n
}

// GetNodes returns the registered nodes
func (e *Engine) GetNodes() map[string]Node {
	return e.nodes
}

// RegisterNodes registers multiple nodes
func (e *Engine) RegisterNodes(nodes map[string]Node) {
	for _, n := range nodes {
		e.nodes[n.ID()] = n
	}
}

// Run executes the workflow
func (e *Engine) Run(ctx context.Context, initialInputs map[string]interface{}) error {
	// Initialize memory with inputs
	for k, v := range initialInputs {
		e.memory.Set(k, v)
	}

	// Build dependency graph
	// Map: NodeID -> []DependentNodeIDs
	adj := make(map[string][]string)
	// Map: NodeID -> InDegree
	inDegree := make(map[string]int)

	// Initialize inDegree for all nodes
	for _, node := range e.workflow.Nodes {
		inDegree[node.ID] = 0
	}

	// Populate graph from edges
	for _, edge := range e.workflow.Edges {
		adj[edge.Source] = append(adj[edge.Source], edge.Target)
		inDegree[edge.Target]++
	}

	// Ready channel for nodes ready to execute
	readyCh := make(chan string, len(e.workflow.Nodes))

	// Find initial nodes (in-degree 0)
	for id, deg := range inDegree {
		if deg == 0 {
			readyCh <- id
		}
	}

	// WaitGroup to wait for all nodes to finish
	var wg sync.WaitGroup
	// Error channel
	errCh := make(chan error, 1)
	// Done channel to signal completion
	doneCh := make(chan struct{})

	// Start a coordinator goroutine
	go func() {
		completedNodes := 0
		totalNodes := len(e.workflow.Nodes)

		for {
			select {
			case <-ctx.Done():
				return
			case nodeID := <-readyCh:
				wg.Add(1)
				go func(id string) {
					defer wg.Done()
					if err := e.executeNode(ctx, id); err != nil {
						select {
						case errCh <- err:
						default:
						}
						return
					}

					// Node finished, update downstream dependencies
					e.mu.Lock()
					completedNodes++

					// Check if node output has a specific branch selected
					var selectedBranch string
					if outputs, ok := e.outputs[id]; ok {
						if val, ok := outputs["_branch_id"]; ok {
							selectedBranch, _ = val.(string)
						}
					}

					for _, neighbor := range adj[id] {
						// Check edge condition
						// We need to find the edge definition to check SourceHandle
						// For MVP efficiency, we should have indexed this, but linear search is fine for now
						var edgeHandle string
						for _, edge := range e.workflow.Edges {
							if edge.Source == id && edge.Target == neighbor {
								edgeHandle = edge.SourceHandle
								break
							}
						}

						// If node selected a branch, only follow edges with matching handle
						// If node didn't select a branch (empty), only follow edges with empty handle
						if selectedBranch == edgeHandle {
							inDegree[neighbor]--
							if inDegree[neighbor] == 0 {
								readyCh <- neighbor
							}
						} else {
							// Branch not taken.
							// In a full implementation, we should propagate "SKIP" to the neighbor.
							// For this MVP, if we don't decrement inDegree, the neighbor will never run.
							// This effectively "skips" it, BUT if the neighbor has other parents that ARE taken,
							// it will wait forever.
							// FIX: We must decrement inDegree but NOT add to readyCh?
							// No, if we decrement, it might run when other parents finish.
							// If we want to SKIP, we need to mark it as skipped and propagate.
							//
							// Simplified Logic for MVP:
							// We assume IfElse nodes are the ONLY parents of their branches in the simple examples.
							// So if we don't trigger it, it won't run.
							// However, to avoid "hanging" the workflow (waiting for all nodes),
							// we should probably count it as "completed" or "skipped".
							//
							// Let's implement a basic "Skip" propagation:
							// If we don't take the branch, we treat the target as "Skipped".
							// We decrement inDegree. If inDegree becomes 0, we check if ANY parent was "Taken".
							// If NO parent was taken (all skipped), then this node is Skipped.
							//
							// FOR MVP HACK:
							// Just decrement inDegree. If it hits 0, check if we should run it.
							// How do we know if we should run it?
							// We need to track if the edge was "activated".
							//
							// Let's refine:
							// We decrement inDegree regardless.
							// But we only add to readyCh if the edge was "activated".
							// Wait, if a node has 2 parents, one activates it, one doesn't.
							// It should run.
							// So, we need to track "activation count" or similar?
							//
							// Simpler approach for MVP:
							// If selectedBranch != edgeHandle, we effectively "skip" this edge.
							// But we MUST decrement inDegree so the graph doesn't stall.
							// But if we decrement inDegree and it hits 0, we put it in readyCh?
							// If we put it in readyCh, it runs. We don't want it to run if it was supposed to be skipped.
							//
							// Let's assume strict branching for MVP:
							// If an edge is NOT taken, we treat the target node as "Skipped" immediately if it depends solely on this edge.
							//
							// Let's go with the "Decrement but don't run if skipped" approach requires state.
							//
							// ALTERNATIVE:
							// If selectedBranch != edgeHandle:
							//   We still decrement inDegree.
							//   If inDegree == 0:
							//     We check if we should run.
							//     How? We look at the inputs?
							//     Or we just run it, and the node itself checks if it has valid inputs?
							//
							// Let's try the simplest valid thing:
							// If edge matches, we process normally.
							// If edge does NOT match, we treat it as "Skipped".
							// We need to propagate this skip.
							//
							// Recursive Skip:
							// func skipNode(nodeID) {
							//   mark node as skipped
							//   completedNodes++
							//   for neighbor in adj[nodeID]:
							//     inDegree[neighbor]--
							//     if inDegree[neighbor] == 0:
							//       skipNode(neighbor)
							// }
							//
							// This seems correct for a DAG.

							e.skipNode(neighbor, adj, inDegree, &completedNodes, &wg)
						}
					}
					isDone := completedNodes == totalNodes
					e.mu.Unlock()

					if isDone {
						close(doneCh)
					}
				}(nodeID)
			case <-doneCh:
				return
			}
		}
	}()

	// Wait for completion or error
	select {
	case <-doneCh:
		wg.Wait() // Ensure all workers finish
		return nil
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (e *Engine) executeNode(ctx context.Context, nodeID string) error {
	// Find node definition
	var nodeDef *dsl.NodeDefinition
	for _, n := range e.workflow.Nodes {
		if n.ID == nodeID {
			nodeDef = &n
			break
		}
	}
	if nodeDef == nil {
		return fmt.Errorf("node definition not found: %s", nodeID)
	}

	// Find node implementation
	nodeImpl, ok := e.nodes[nodeID]
	if !ok {
		return fmt.Errorf("node implementation not found: %s", nodeID)
	}

	// Resolve inputs
	inputs := make(map[string]interface{})
	for k, v := range nodeDef.Inputs {
		val, err := e.resolveValue(v)
		if err != nil {
			return fmt.Errorf("failed to resolve input %s for node %s: %w", k, nodeID, err)
		}
		inputs[k] = val
	}

	// Execute
	fmt.Printf("Executing node: %s (Type: %s)\n", nodeID, nodeDef.Type)
	outputs, err := nodeImpl.Execute(&NodeContext{
		Ctx:    ctx,
		Memory: e.memory,
		Inputs: inputs,
		NodeID: nodeID,
		Engine: e,
	})
	if err != nil {
		return fmt.Errorf("node execution failed: %w", err)
	}

	// Store outputs
	e.mu.Lock()
	e.outputs[nodeID] = outputs
	e.mu.Unlock()

	return nil
}

func (e *Engine) resolveValue(template string) (interface{}, error) {
	// Check if it's a pure template "{{ key }}" -> return raw value (could be map/list)
	trimmed := strings.TrimSpace(template)
	if strings.HasPrefix(trimmed, "{{") && strings.HasSuffix(trimmed, "}}") {
		// It might be a pure template, but we need to check if there are other chars
		// For MVP, strict check:
		key := strings.TrimSpace(trimmed[2 : len(trimmed)-2])
		// If key contains "}}", it's likely multiple templates or invalid, but let's assume simple key
		return e.resolveKey(key)
	}

	// Mixed template "Hello {{ name }}" -> return string
	// Find all {{ ... }}
	start := 0
	var sb strings.Builder
	for {
		open := strings.Index(template[start:], "{{")
		if open == -1 {
			sb.WriteString(template[start:])
			break
		}
		open += start
		relClose := strings.Index(template[open:], "}}")
		if relClose == -1 {
			// Malformed, just return rest
			sb.WriteString(template[start:])
			break
		}
		close := open + relClose + 2 // absolute index of end of }}

		// Append text before
		sb.WriteString(template[start:open])

		// Resolve key
		key := strings.TrimSpace(template[open+2 : close-2])
		val, err := e.resolveKey(key)
		if err != nil {
			return nil, err
		}
		sb.WriteString(fmt.Sprintf("%v", val))

		start = close
	}
	return sb.String(), nil
}

func (e *Engine) resolveKey(key string) (interface{}, error) {
	// Check memory
	if strings.HasPrefix(key, "memory.") {
		memKey := strings.TrimPrefix(key, "memory.")
		val, ok := e.memory.Get(memKey)
		if !ok {
			return nil, fmt.Errorf("memory key not found: %s", memKey)
		}
		return val, nil
	}

	// Check node output: node_id.output_key
	parts := strings.Split(key, ".")
	if len(parts) == 2 {
		nodeID := parts[0]
		outputKey := parts[1]

		e.mu.RLock()
		nodeOutputs, ok := e.outputs[nodeID]
		e.mu.RUnlock()

		if !ok {
			return nil, fmt.Errorf("node outputs not found: %s", nodeID)
		}
		val, ok := nodeOutputs[outputKey]
		if !ok {
			return nil, fmt.Errorf("output key %s not found in node %s", outputKey, nodeID)
		}
		return val, nil
	}

	return nil, fmt.Errorf("invalid template key: %s", key)
}

func (e *Engine) skipNode(nodeID string, adj map[string][]string, inDegree map[string]int, completedNodes *int, wg *sync.WaitGroup) {
	// This node is skipped.
	// In a real implementation, we might want to record this state.
	fmt.Printf("Skipping node: %s\n", nodeID)

	(*completedNodes)++

	// Propagate skip to neighbors
	for _, neighbor := range adj[nodeID] {
		inDegree[neighbor]--
		if inDegree[neighbor] == 0 {
			e.skipNode(neighbor, adj, inDegree, completedNodes, wg)
		}
	}
}
