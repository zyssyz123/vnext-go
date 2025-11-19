# Dify vNext Go Examples

This directory contains example workflows demonstrating the capabilities of the `dify-vnext-go` engine.

## Running Examples

You can run any example using the following command:

```bash
go run cmd/main.go -f examples/<filename>.yaml
```

## Available Examples

### 1. Simple & Research
- **`simple.yaml`**: A basic linear workflow (Start -> LLM -> End).
- **`research.yaml`**: A complex workflow demonstrating loops, conditional branching, and code execution.

### 2. Customer Support Triage (`support_triage.yaml`)
Demonstrates **Conditional Routing** based on user intent.
- **Flow**: Classifies a ticket -> Routes to Billing, Technical, or General support -> Generates response.
- **Key Features**: `IfElse` node with multiple cases, template resolution for skipped nodes.

### 3. Automated Code Review (`code_review.yaml`)
Demonstrates **Parallel Execution** and **Multi-Perspective Analysis**.
- **Flow**: Analyzes code for Security, Style, and Performance in parallel -> Aggregates results using Code node.
- **Key Features**: Parallel LLM calls, complex inputs (multi-line code snippet).

### 4. Multi-language Translation (`translation.yaml`)
Demonstrates **Looping** over a list of items.
- **Flow**: Takes a text and a list of target languages -> Loops over languages to translate -> Formats results.
- **Key Features**: `Loop` node, sub-workflow execution, list handling.

## Notes
- Ensure you have `OPENAI_API_KEY` set in your environment for actual LLM calls. Otherwise, the engine will use mock responses.
