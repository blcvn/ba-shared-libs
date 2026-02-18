package utils

import (
	"context"
	"fmt"

	proxy_pb "github.com/blcvn/kratos-proto/go/ai-proxy"
	prompt_pb "github.com/blcvn/kratos-proto/go/prompt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ServiceBridge struct {
	promptClient  prompt_pb.PromptServiceClient
	aiProxyClient proxy_pb.AIProxyServiceClient
}

func NewServiceBridge(promptAddr, aiProxyAddr string) (*ServiceBridge, error) {
	pConn, err := grpc.Dial(promptAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	aConn, err := grpc.Dial(aiProxyAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &ServiceBridge{
		promptClient:  prompt_pb.NewPromptServiceClient(pConn),
		aiProxyClient: proxy_pb.NewAIProxyServiceClient(aConn),
	}, nil
}

func (b *ServiceBridge) RenderPrompt(ctx context.Context, name string, variables map[string]string) (string, error) {
	resp, err := b.promptClient.RenderTemplate(ctx, &prompt_pb.RenderTemplateRequest{
		Payload: &prompt_pb.RenderTemplatePayload{
			TemplateId: name, // Using name as ID for convenience in rendering
			Variables:  variables,
		},
	})
	if err != nil {
		return "", err
	}
	return resp.Rendered.RenderedText, nil
}

func (b *ServiceBridge) Complete(ctx context.Context, modelID, systemPrompt, userPrompt string) (string, error) {
	resp, err := b.aiProxyClient.Complete(ctx, &proxy_pb.CompleteRequest{
		Payload: &proxy_pb.CompletePayload{
			ModelId: modelID,
			Messages: []*proxy_pb.ChatMessage{
				{Role: proxy_pb.ChatMessage_SYSTEM, Content: systemPrompt},
				{Role: proxy_pb.ChatMessage_USER, Content: userPrompt},
			},
		},
	})
	if err != nil {
		return "", err
	}
	if resp.Result.Code != proxy_pb.ResultCode_SUCCESS {
		return "", fmt.Errorf("ai proxy error: %s", resp.Result.Message)
	}
	return resp.Completion.Text, nil
}
