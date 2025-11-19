package nodes

import (
	"dify-vnext-go/pkg/engine"
	"fmt"
	"io"
	"net/http"
)

type HttpRequestNode struct {
	BaseNode
	Method string
	URL    string
}

func NewHttpRequestNode(id string, config map[string]interface{}) *HttpRequestNode {
	method, _ := config["method"].(string)
	url, _ := config["url"].(string)
	if method == "" {
		method = "GET"
	}
	return &HttpRequestNode{
		BaseNode: NewBaseNode(id, "HttpRequest"),
		Method:   method,
		URL:      url,
	}
}

func (n *HttpRequestNode) Execute(ctx *engine.NodeContext) (map[string]interface{}, error) {
	// Resolve URL if it contains templates (simple implementation)
	// In a real implementation, we'd resolve all config fields.
	// For MVP, let's assume URL might be passed in inputs if not in config,
	// or we just use the config.

	url := n.URL
	// Check if URL is dynamic from inputs
	if val, ok := ctx.Inputs["url"]; ok {
		url = fmt.Sprintf("%v", val)
	}

	fmt.Printf("[%s] Making HTTP Request: %s %s\n", n.ID(), n.Method, url)

	req, err := http.NewRequest(n.Method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers if needed (omitted for MVP)

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

	fmt.Printf("[%s] Response Status: %s\n", n.ID(), resp.Status)

	return map[string]interface{}{
		"status_code": resp.StatusCode,
		"body":        string(body),
	}, nil
}
