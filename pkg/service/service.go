package service

import (
	"context"
	"time"
)

// FileRequest struct contains file handler request data
type FileRequest struct {
	Agent      string                 `json:"agent"`
	Action     string                 `json:"action"`
	Parameters map[string]interface{} `json:"parameters"`
	Files      []string               `json:"files"`
}

type Message struct {
	ID     string `json:"id"`
	Result Result `json:"result"`
}

type Result struct {
	OutputFiles        []string `json:"output_files"`
	RawProcessorOutput string   `json:"raw_processor_output"`
	MetaData           []string `json:"metadata"`
	ProcessingTime     string   `json:"processing_time"`
}

// FileResponse struct file response data
type FileResponse struct {
	Success bool    `json:"success"`
	Message Message `json:"message"`
	Error   string  `json:"error,omitempty"`
}

// HealthResponse struct health response data
type HealthResponse struct {
	Status  string    `json:"status"`
	Time    time.Time `json:"time"`
	Version string    `json:"version"`
}

// FileHandlerService file handler method interface
type FileHandlerService interface {
	ProcessFile(ctx context.Context, req FileRequest) (FileResponse, error)
	Health(ctx context.Context) (HealthResponse, error)
}
