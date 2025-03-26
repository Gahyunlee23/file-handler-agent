package service

import (
	"context"
	"errors"
	"file-handler-agent/pkg/service/agent"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// fileHandlerService implement interface FileHandlerService
type fileHandlerService struct {
	agentRegistry agent.Registry
}

// NewFileHandlerService generate new FileHandlerService instance
func NewFileHandlerService(registry agent.Registry) FileHandlerService {
	return &fileHandlerService{
		agentRegistry: registry,
	}
}

// ProcessFile process the file based on the request
func (s *fileHandlerService) ProcessFile(ctx context.Context, req FileRequest) (FileResponse, error) {
	// find an agent
	agentImpl, exists := s.agentRegistry.Get(req.Agent)
	if !exists {
		return FileResponse{
			Success: false,
			Error:   fmt.Sprintf("agent '%s' not found", req.Agent),
		}, errors.New("agent not found")
	}

	// Generate a unique ID for this processing request
	requestID := generateUniqueID()

	// Create a directory for this request if it doesn't exist
	outputDir := filepath.Join("temp", "output", requestID)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return FileResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to create output directory: %v", err),
		}, err
	}

	// Add the requestID to parameters so the agent can use it
	if req.Parameters == nil {
		req.Parameters = make(map[string]interface{})
	}

	req.Parameters["request_id"] = requestID
	req.Parameters["output_dir"] = outputDir

	// execute agent
	outputFiles, err := agentImpl.Execute(ctx, req.Action, req.Parameters, req.Files)
	if err != nil {
		return FileResponse{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	log.Printf("outputFiles: %v", outputFiles)

	// Get the raw processor output if available
	rawOutput := ""
	if procOutput, ok := req.Parameters["processorOutput"].(string); ok {
		rawOutput = procOutput
	}

	// Create response with the new format
	return FileResponse{
		Success: true,
		Message: Message{
			ID: requestID,
			Result: Result{
				OutputFiles:        outputFiles,
				RawProcessorOutput: rawOutput,
				MetaData:           []string{},
			},
		},
	}, nil
}

func generateUniqueID() string {
	return fmt.Sprintf("%x", time.Now().UnixNano())
}

// Health check the condition
func (s *fileHandlerService) Health(ctx context.Context) (HealthResponse, error) {
	return HealthResponse{
		Status:  "OK",
		Time:    time.Now(),
		Version: "1.0.0",
	}, nil
}
