package nodes

import (
	"dify-vnext-go/pkg/engine"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/dop251/goja"
)

type ToolNode struct {
	BaseNode
	ProviderID string
	ToolID     string
}

func NewToolNode(id string, config map[string]interface{}) *ToolNode {
	provider, _ := config["provider_id"].(string)
	tool, _ := config["tool_id"].(string)
	return &ToolNode{
		BaseNode:   NewBaseNode(id, "Tool"),
		ProviderID: provider,
		ToolID:     tool,
	}
}

func (n *ToolNode) Execute(ctx *engine.NodeContext) (map[string]interface{}, error) {
	fmt.Printf("[%s] Executing Tool: %s/%s\n", n.ID(), n.ProviderID, n.ToolID)

	switch n.ToolID {
	case "google_search":
		return n.executeGoogleSearch(ctx)
	case "calculator":
		return n.executeCalculator(ctx)
	default:
		return nil, fmt.Errorf("unknown tool: %s", n.ToolID)
	}
}

func (n *ToolNode) executeCalculator(ctx *engine.NodeContext) (map[string]interface{}, error) {
	expression, _ := ctx.Inputs["expression"].(string)
	if expression == "" {
		return nil, fmt.Errorf("missing input 'expression'")
	}

	fmt.Printf("[%s] Calculating: %s\n", n.ID(), expression)

	vm := goja.New()
	val, err := vm.RunString(expression)
	if err != nil {
		return nil, fmt.Errorf("calculation failed: %w", err)
	}

	result := val.String()
	fmt.Printf("[%s] Calculation Result: %s\n", n.ID(), result)

	return map[string]interface{}{
		"text": result,
	}, nil
}

type SerpApiResponse struct {
	OrganicResults []struct {
		Title   string `json:"title"`
		Link    string `json:"link"`
		Snippet string `json:"snippet"`
	} `json:"organic_results"`
	Error string `json:"error"`
}

func (n *ToolNode) executeGoogleSearch(ctx *engine.NodeContext) (map[string]interface{}, error) {
	query, _ := ctx.Inputs["query"].(string)
	if query == "" {
		return nil, fmt.Errorf("missing input 'query'")
	}

	apiKey := os.Getenv("SERPAPI_API_KEY")
	if apiKey == "" {
		fmt.Printf("[%s] WARNING: SERPAPI_API_KEY not set. Using Mock response.\n", n.ID())
		return map[string]interface{}{
			"text": fmt.Sprintf("Mock Search Results for '%s': [Real Search requires API Key]", query),
		}, nil
	}

	fmt.Printf("[%s] Searching Google via SerpApi: %s\n", n.ID(), query)

	u, _ := url.Parse("https://serpapi.com/search")
	q := u.Query()
	q.Set("q", query)
	q.Set("api_key", apiKey)
	q.Set("engine", "google")
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	var serpResp SerpApiResponse
	if err := json.Unmarshal(body, &serpResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if serpResp.Error != "" {
		return nil, fmt.Errorf("SerpApi error: %s", serpResp.Error)
	}

	// Format results
	var resultText string
	for i, res := range serpResp.OrganicResults {
		if i >= 3 {
			break
		}
		resultText += fmt.Sprintf("%d. %s: %s\n", i+1, res.Title, res.Snippet)
	}

	if resultText == "" {
		resultText = "No results found."
	}

	fmt.Printf("[%s] Search Results found.\n", n.ID())

	return map[string]interface{}{
		"text": resultText,
	}, nil
}
