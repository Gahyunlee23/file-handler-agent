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

// FileResponse struct file response data
type FileResponse struct {
	Success bool     `json:"success"`
	Files   []string `json:"files"`
	Error   string   `json:"error,omitempty"`
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
