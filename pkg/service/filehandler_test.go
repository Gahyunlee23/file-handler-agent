package service

import (
	"context"
	"file-handler-agent/pkg/service/agent"
	"testing"
)

type MockAgent struct {
	ExecuteFn      func(ctx context.Context, action string, params map[string]interface{}, files []string) ([]string, error)
	ExecuteInvoked bool
}

func (m *MockAgent) Execute(ctx context.Context, action string, params map[string]interface{}, files []string) ([]string, error) {
	m.ExecuteInvoked = true
	return m.ExecuteFn(ctx, action, params, files)
}

func TestProcessFile(t *testing.T) {
	testCases := []struct {
		name          string
		req           FileRequest
		mockAgent     *MockAgent
		expectedFiles []string
		expectError   bool
	}{
		{
			name: "Success case",
			req: FileRequest{
				Agent:      "mock",
				Action:     "testAction",
				Parameters: map[string]interface{}{"param1": "value1"},
				Files:      []string{"file1.pdf"},
			},
			mockAgent: &MockAgent{
				ExecuteFn: func(ctx context.Context, action string, params map[string]interface{}, files []string) ([]string, error) {
					return []string{"output1.png"}, nil
				},
			},
			expectedFiles: []string{"output1.png"},
			expectError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			registry := agent.NewRegistry()
			registry.Register("mock", tc.mockAgent)

			svc := NewFileHandlerService(registry)

			resp, err := svc.ProcessFile(context.Background(), tc.req)

			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tc.expectError {
				if !resp.Success {
					t.Error("Expected success to be true")
				}
				if len(resp.Files) != len(tc.expectedFiles) {
					t.Errorf("Expected %d files, got %d", len(tc.expectedFiles), len(resp.Files))
				}
			}

			if !tc.mockAgent.ExecuteInvoked {
				t.Error("Expected agent.Execute to be called")
			}
		})
	}
}
