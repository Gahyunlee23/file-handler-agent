package service

import (
	"context"
	"errors"
	"file-handler-agent/pkg/service/agent"
	"fmt"
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

	// execute agent
	outputFiles, err := agentImpl.Execute(ctx, req.Action, req.Parameters, req.Files)
	if err != nil {
		return FileResponse{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// return success
	return FileResponse{
		Success: true,
		Files:   outputFiles,
	}, nil
}

// Health check the condition
func (s *fileHandlerService) Health(ctx context.Context) (HealthResponse, error) {
	return HealthResponse{
		Status:  "OK",
		Time:    time.Now(),
		Version: "1.0.0",
	}, nil
}
