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
	startTime := time.Now()
	log.Printf("Starting file processing request with agent: %s, action: %s", req.Agent, req.Action)

	// Check for context cancellation early
	select {
	case <-ctx.Done():
		return FileResponse{
			Success: false,
			Error:   ctx.Err().Error(),
		}, ctx.Err()
	default:
		// Continue with processing
	}

	// Find an agent
	agentImpl, exists := s.agentRegistry.Get(req.Agent)
	if !exists {
		log.Printf("Agent not found: %s", req.Agent)
		return FileResponse{
			Success: false,
			Error:   fmt.Sprintf("agent '%s' not found", req.Agent),
		}, errors.New("agent not found")
	}

	// Generate a unique ID for this processing request
	requestID := generateUniqueID()
	log.Printf("Generated request ID: %s", requestID)

	// Add the requestID to parameters
	if req.Parameters == nil {
		req.Parameters = make(map[string]interface{})
	}

	// Generate output directory path but don't create it yet
	// The directory will be created by the agent after validation
	outputDir := filepath.Join("temp", "output", requestID)
	req.Parameters["request_id"] = requestID
	req.Parameters["output_dir"] = outputDir

	// Check if cleanup is requested
	cleanupTemp, _ := req.Parameters["cleanup_temp"].(bool)

	// Setup deferred cleanup if requested
	if cleanupTemp {
		defer func() {
			// Schedule cleanup with a small delay to ensure files are fully processed
			go func() {
				cleanupTime := 5 * time.Minute
				log.Printf("Scheduling cleanup of directory %s in %v", outputDir, cleanupTime)
				time.Sleep(cleanupTime)
				log.Printf("Cleaning up temporary directory: %s", outputDir)
				if err := os.RemoveAll(outputDir); err != nil {
					log.Printf("Failed to clean up temp directory: %v", err)
				}
			}()
		}()
	}

	// Execute agent with timeout if specified
	var outputFiles []string
	var err error

	timeout, ok := req.Parameters["timeout"].(float64)
	if ok && timeout > 0 {
		// Create a timeout context
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()

		log.Printf("Executing agent with timeout: %.2f seconds", timeout)
		outputFiles, err = agentImpl.Execute(timeoutCtx, req.Action, req.Parameters, req.Files)
	} else {
		// Use original context
		log.Printf("Executing agent without timeout")
		outputFiles, err = agentImpl.Execute(ctx, req.Action, req.Parameters, req.Files)
	}

	if err != nil {
		log.Printf("Agent execution failed: %v", err)
		return FileResponse{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	log.Printf("Agent execution successful. Output files: %v", outputFiles)

	// Get the raw processor output if available
	rawOutput := ""
	if procOutput, ok := req.Parameters["processorOutput"].(string); ok {
		rawOutput = procOutput
	}

	// Add any metadata information
	var metadata []string
	if metaInfo, ok := req.Parameters["metadata"].([]string); ok {
		metadata = metaInfo
	}

	// Log performance metrics
	processingTime := time.Since(startTime)
	log.Printf("Request %s completed in %v", requestID, processingTime)

	// Create response with the new format
	return FileResponse{
		Success: true,
		Message: Message{
			ID: requestID,
			Result: Result{
				OutputFiles:        outputFiles,
				RawProcessorOutput: rawOutput,
				MetaData:           metadata,
				ProcessingTime:     processingTime.String(),
			},
		},
	}, nil
}

// generateUniqueID creates a unique identifier for a request
func generateUniqueID() string {
	return fmt.Sprintf("%x", time.Now().UnixNano())
}

// CleanupTemporaryFiles removes old temporary directories
func (s *fileHandlerService) CleanupTemporaryFiles(olderThan time.Duration) error {
	baseDir := filepath.Join("temp", "output")
	cutoffTime := time.Now().Add(-olderThan)

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirPath := filepath.Join(baseDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			log.Printf("Failed to get info for %s: %v", dirPath, err)
			continue
		}

		if info.ModTime().Before(cutoffTime) {
			log.Printf("Removing old temp directory: %s (modified: %v)",
				dirPath, info.ModTime())
			if err := os.RemoveAll(dirPath); err != nil {
				log.Printf("Failed to remove directory %s: %v", dirPath, err)
			}
		}
	}

	return nil
}

// Health check the condition
func (s *fileHandlerService) Health(ctx context.Context) (HealthResponse, error) {
	return HealthResponse{
		Status:  "OK",
		Time:    time.Now(),
		Version: "1.0.0",
	}, nil
}
