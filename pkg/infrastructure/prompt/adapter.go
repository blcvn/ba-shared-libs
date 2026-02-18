package prompt

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	pb "github.com/blcvn/kratos-proto/go/prompt"
)

// PromptAdapter adapts the Prompt Service to a simpler usage interface.
type PromptAdapter struct {
	client      pb.PromptServiceClient
	templateDir string
}

func NewPromptAdapter(client pb.PromptServiceClient, templateDir string) *PromptAdapter {
	return &PromptAdapter{
		client:      client,
		templateDir: templateDir,
	}
}

// GetPrompt fetches a prompt by name and replaces variables.
// In PRD-URD, this was loaded from file. Here we call the service.
// If service fails, it falls back to local template files.
func (p *PromptAdapter) GetPrompt(ctx context.Context, name string, vars map[string]string) (string, error) {
	// Name conversion: "reasoning.txt" -> "reasoning" (if needed by service)
	// Assuming service takes the prompt name as identifier.
	// Clean extension if service doesn't expect it.
	promptName := strings.TrimSuffix(name, ".txt")

	req := &pb.RenderTemplateRequest{
		Payload: &pb.RenderTemplatePayload{
			TemplateId: promptName, // Service treats TemplateId as Name
			Variables:  vars,
		},
	}

	resp, err := p.client.RenderTemplate(ctx, req)
	if err == nil && resp.Result != nil && resp.Result.Code == pb.ResultCode_SUCCESS {
		return resp.Rendered.RenderedText, nil
	}

	// Fallback to local template
	log.Printf("WARN: Prompt service failed for '%s' (err: %v), falling back to local template", name, err)
	return p.loadLocalPrompt(name, vars)
}

func (p *PromptAdapter) loadLocalPrompt(name string, vars map[string]string) (string, error) {
	if !strings.HasSuffix(name, ".txt") {
		name += ".txt"
	}
	path := filepath.Join(p.templateDir, name)
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to load local prompt '%s' from %s: %w", name, path, err)
	}

	rendered := string(content)
	for k, v := range vars {
		rendered = strings.ReplaceAll(rendered, k, v)
	}

	return rendered, nil
}
