package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	aiproxy "github.com/blcvn/kratos-proto/go/ai-proxy"
)

// AIProxyAdapter implements the domain.LLMService interface using the AI Proxy Service.
type AIProxyAdapter struct {
	client  aiproxy.AIProxyServiceClient
	modelID string
}

func NewAIProxyAdapter(client aiproxy.AIProxyServiceClient, modelID string) *AIProxyAdapter {
	return &AIProxyAdapter{
		client:  client,
		modelID: modelID,
	}
}

// Chat sends a chat request to the AI Proxy.
func (a *AIProxyAdapter) Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	// Construct the prompt.
	// Note: AI Proxy 'Prompt' field in CompleteRequest is usually a single string.
	// We combine system and user prompt.
	fullPrompt := ""
	if systemPrompt != "" {
		fullPrompt = fmt.Sprintf("%s\n\n%s", systemPrompt, userPrompt)
	} else {
		fullPrompt = userPrompt
	}

	req := &aiproxy.CompleteRequest{
		Payload: &aiproxy.CompletePayload{
			ModelId:     a.modelID,
			Prompt:      fullPrompt,
			Temperature: 0.2, // Low temp for deterministic logic
			MaxTokens:   4096,
		},
	}

	// Debug Log
	reqJSON, _ := json.Marshal(req)
	log.Printf("AIProxyAdapter Chat Request: %s", string(reqJSON))

	// Use specific timeout if ctx doesn't have one?
	// The grpc client usually handles context timeout.
	resp, err := a.client.Complete(ctx, req)
	if err != nil {
		return "", fmt.Errorf("AI Proxy call failed: %w", err)
	}

	if resp.Result.Code != aiproxy.ResultCode_SUCCESS {
		return "", fmt.Errorf("AI Proxy returned error: %s", resp.Result.Message)
	}

	return resp.Completion.Text, nil
}
