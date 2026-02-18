package prompt

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	pb "github.com/blcvn/kratos-proto/go/prompt"
	"google.golang.org/grpc"
)

// MockPromptClient is a mock for pb.PromptServiceClient
type MockPromptClient struct {
	pb.PromptServiceClient
	ShouldFail bool
	Response   *pb.RenderTemplateResponse
}

func (m *MockPromptClient) RenderTemplate(ctx context.Context, in *pb.RenderTemplateRequest, opts ...grpc.CallOption) (*pb.RenderTemplateResponse, error) {
	if m.ShouldFail {
		return nil, os.ErrNotExist // Generic error
	}
	return m.Response, nil
}

func TestGetPrompt_Fallback(t *testing.T) {
	// 1. Setup temporary template directory
	tmpDir, err := ioutil.TempDir("", "templates")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	promptContent := "Hello {{NAME}}!"
	err = ioutil.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte(promptContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// 2. Test Success from Service
	mockClient := &MockPromptClient{
		ShouldFail: false,
		Response: &pb.RenderTemplateResponse{
			Result: &pb.Result{Code: pb.ResultCode_SUCCESS},
			Rendered: &pb.RenderedPrompt{
				RenderedText: "Rendered by Service",
			},
		},
	}
	adapter := NewPromptAdapter(mockClient, tmpDir)

	vars := map[string]string{"{{NAME}}": "World"}
	res, err := adapter.GetPrompt(context.Background(), "test.txt", vars)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if res != "Rendered by Service" {
		t.Errorf("Expected 'Rendered by Service', got '%s'", res)
	}

	// 3. Test Fallback on Service Failure
	mockClient.ShouldFail = true
	res, err = adapter.GetPrompt(context.Background(), "test.txt", vars)
	if err != nil {
		t.Errorf("Unexpected error on fallback: %v", err)
	}
	if res != "Hello World!" {
		t.Errorf("Expected 'Hello World!', got '%s'", res)
	}

	// 4. Test Fallback on Service Result Error
	mockClient.ShouldFail = false
	mockClient.Response.Result.Code = 1 // Any non-zero value for error fallback
	res, err = adapter.GetPrompt(context.Background(), "test.txt", vars)
	if err != nil {
		t.Errorf("Unexpected error on fallback: %v", err)
	}
	if res != "Hello World!" {
		t.Errorf("Expected 'Hello World!', got '%s'", res)
	}
}
