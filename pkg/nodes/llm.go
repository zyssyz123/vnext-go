package nodes

import (
	"bytes"
	"dify-vnext-go/pkg/engine"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type LLMNode struct {
	BaseNode
	Model string
}

func NewLLMNode(id string, config map[string]interface{}) *LLMNode {
	model, _ := config["model"].(string)
	if model == "" {
		model = "gpt-3.5-turbo"
	}
	return &LLMNode{
		BaseNode: NewBaseNode(id, "LLM"),
		Model:    model,
	}
}

type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (n *LLMNode) Execute(ctx *engine.NodeContext) (map[string]interface{}, error) {
	prompt, _ := ctx.Inputs["prompt"].(string)
	fmt.Printf("[%s] Calling OpenAI (%s) with prompt: %s\n", n.ID(), n.Model, prompt)

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		// Fallback to mock if no key provided, for safety/testing without cost
		fmt.Printf("[%s] WARNING: OPENAI_API_KEY not set. Using Mock response.\n", n.ID())
		return map[string]interface{}{
			"response": fmt.Sprintf("Mock response (No Key) from %s: %s", n.Model, prompt),
		}, nil
	}

	reqBody := OpenAIRequest{
		Model: n.Model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var openAIResp OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from OpenAI")
	}

	content := openAIResp.Choices[0].Message.Content
	fmt.Printf("[%s] OpenAI Response: %s...\n", n.ID(), content[:min(len(content), 50)])

	return map[string]interface{}{
		"response": content,
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
