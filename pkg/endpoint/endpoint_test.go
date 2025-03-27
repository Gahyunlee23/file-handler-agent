package endpoint

import (
	"context"
	"file-handler-agent/pkg/service"
	"testing"
	"time"
)

type MockService struct {
	ProcessFileFn func(ctx context.Context, req service.FileRequest) (service.FileResponse, error)
	HealthFn      func(ctx context.Context) (service.HealthResponse, error)
}

func (m *MockService) ProcessFile(ctx context.Context, req service.FileRequest) (service.FileResponse, error) {
	return m.ProcessFileFn(ctx, req)
}

func (m *MockService) Health(ctx context.Context) (service.HealthResponse, error) {
	return m.HealthFn(ctx)
}

func TestProcessFileEndpoint(t *testing.T) {
	mockSvc := &MockService{
		ProcessFileFn: func(ctx context.Context, req service.FileRequest) (service.FileResponse, error) {
			return service.FileResponse{
				Success: true,
				Message: service.Message{
					ID: "67e452be630a0",
					Result: service.Result{
						OutputFiles: []string{"output/67e452be630a0.png"},
					},
				},
			}, nil
		},
	}

	endpoint := MakeProcessFileEndpoint(mockSvc)

	req := service.FileRequest{
		Agent:  "ghostscript",
		Action: "convertPdfToImage",
		Parameters: map[string]interface{}{
			"resolution": 72.0,
		},
		Files: []string{"input.pdf"},
	}

	resp, err := endpoint(context.Background(), req)

	// validate
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	fileResp, ok := resp.(service.FileResponse)
	if !ok {
		t.Fatal("Expected FileResponse type")
	}

	if !fileResp.Success {
		t.Error("Expected success to be true")
	}

	if len(fileResp.Message.Result.OutputFiles) != 1 || fileResp.Message.Result.OutputFiles[0] != "output/67e452be630a0.png" {
		t.Errorf("Unexpected files in response: %v", fileResp.Message.Result.OutputFiles)
	}
}

func TestHealthEndpoint(t *testing.T) {
	mockSvc := &MockService{
		HealthFn: func(ctx context.Context) (service.HealthResponse, error) {
			return service.HealthResponse{
				Status:  "OK",
				Time:    time.Now(),
				Version: "1.0.0",
			}, nil
		},
	}

	endpoint := MakeHealthEndpoint(mockSvc)

	resp, err := endpoint(context.Background(), nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	healthResp, ok := resp.(service.HealthResponse)
	if !ok {
		t.Fatal("Expected HealthResponse type")
	}

	if healthResp.Status != "OK" {
		t.Errorf("Expected status OK, got %s", healthResp.Status)
	}

	if healthResp.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", healthResp.Version)
	}
}
