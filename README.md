# Dify vNext (Go)

**Dify vNext** is a high-performance, concurrent workflow engine written in Go. It is designed as a next-generation runtime for LLM-based applications, addressing the performance and architectural limitations of the original Python-based Dify engine.

## 🚀 Key Features

*   **Native Concurrency**: Built on Go's Goroutines and Channels. Supports massive parallelism (e.g., thousands of concurrent loop iterations) with minimal overhead.
*   **Hierarchical Memory System**: Implements a tree-based scoping mechanism. Each loop iteration or parallel branch gets its own isolated memory scope, preventing variable pollution and ensuring thread safety.
*   **State Checkpointing**: Inspired by LangGraph, the engine supports a `Checkpointer` interface. State is saved as immutable snapshots after every node execution, enabling future features like "Time Travel" and reliable retries.
*   **Type-Safe & Compiled**: Static typing catches errors at build time. Deploys as a single, lightweight binary.

## 🏗️ Architecture

```mermaid
graph TB
    subgraph "Workflow Definition"
        YAML[YAML Workflow File]
    end
    
    subgraph "DSL Layer"
        Parser[DSL Parser]
        WorkflowDef[WorkflowDefinition]
    end
    
    subgraph "Core Engine"
        Engine[Engine Runtime]
        Memory[Hierarchical Memory]
        Checkpointer[Checkpointer Interface]
        
        subgraph "Memory Scopes"
            RootScope[Root Scope]
            ChildScope1[Child Scope 1]
            ChildScope2[Child Scope 2]
            RootScope -.parent.-> ChildScope1
            RootScope -.parent.-> ChildScope2
        end
    end
    
    subgraph "Node Implementations"
        StartNode[Start Node]
        LLMNode[LLM Node]
        CodeNode[Code Node]
        LoopNode[Loop Node]
        IfElseNode[IfElse Node]
        ToolNode[Tool Node]
        AnswerNode[Answer Node]
    end
    
    subgraph "State Persistence"
        InMemoryCP[InMemoryCheckpointer]
        FutureCP[Redis/Postgres<br/>Checkpointer]
    end
    
    YAML --> Parser
    Parser --> WorkflowDef
    WorkflowDef --> Engine
    
    Engine --> Memory
    Engine --> Checkpointer
    Engine -.executes.-> StartNode
    Engine -.executes.-> LLMNode
    Engine -.executes.-> CodeNode
    Engine -.executes.-> LoopNode
    Engine -.executes.-> IfElseNode
    Engine -.executes.-> ToolNode
    Engine -.executes.-> AnswerNode
    
    Checkpointer -.implements.-> InMemoryCP
    Checkpointer -.future.-> FutureCP
    
    LoopNode -.creates.-> ChildScope1
    LoopNode -.creates.-> ChildScope2
    
    style Engine fill:#4A90E2,stroke:#2E5C8A,stroke-width:3px,color:#fff
    style Memory fill:#50C878,stroke:#2E7D4E,stroke-width:2px,color:#fff
    style Checkpointer fill:#F39C12,stroke:#C87F0A,stroke-width:2px,color:#fff
```

## 📂 Project Structure

```
dify-vnext-go/
├── cmd/
│   └── main.go           # Application entry point
├── pkg/
│   ├── dsl/              # Workflow DSL definitions and YAML parser
│   ├── engine/           # Core runtime (Engine, Memory, State/Checkpointer)
│   └── nodes/            # Node implementations (Start, LLM, Code, Loop, etc.)
├── examples/             # Example workflow YAML files
└── go.mod                # Go module definition
```

## 🛠️ Getting Started

### Prerequisites

*   **Go 1.21+** installed.
*   **OpenAI API Key** (Optional, for LLM nodes).
    ```bash
    export OPENAI_API_KEY="sk-..."
    ```
    *If not provided, LLM nodes will return mock responses.*

### Running Examples

The project comes with several example workflows to demonstrate its capabilities.

1.  **Simple Workflow** (Linear execution):
    ```bash
    go run cmd/main.go -f examples/simple.yaml
    ```

2.  **Complex Workflow** (Branching & Tools):
    ```bash
    go run cmd/main.go -f examples/complex.yaml
    ```

3.  **Customer Support Triage** (Conditional Routing):
    ```bash
    go run cmd/main.go -f examples/support_triage.yaml
    ```

4.  **Automated Code Review** (Parallel Execution):
    ```bash
    go run cmd/main.go -f examples/code_review.yaml
    ```

5.  **Multi-language Translation** (Loops):
    ```bash
    go run cmd/main.go -f examples/translation.yaml
    ```

## 🧠 Architecture Highlights

### Memory Management
Unlike Dify's flat variable pool, vNext uses **Hierarchical Scoping**.
- **Global Scope**: Inputs to the `Start` node.
- **Child Scope**: Created for each `Loop` iteration.
- **Bubble-Up Lookup**: Variables are looked up in the current scope, then the parent, up to the root.

### Checkpointing
The engine integrates a `Checkpointer` that saves the state of the entire memory tree after each node execution.
- **Current Implementation**: `InMemoryCheckpointer` (for MVP/Testing).
- **Future**: Redis/Postgres implementations for persistent state.

## 🤝 Contributing
Contributions are welcome! Please check the `pkg/nodes` directory to see how to implement new node types.
