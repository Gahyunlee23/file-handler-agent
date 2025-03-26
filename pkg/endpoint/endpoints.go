package endpoint

import (
	"context"
	errors "file-handler-agent/pkg/error"
	"file-handler-agent/pkg/service"

	"github.com/go-kit/kit/endpoint"
)

// MakeProcessFileEndpoint ProcessFile service endpoint
func MakeProcessFileEndpoint(svc service.FileHandlerService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(service.FileRequest)
		if !ok {
			return nil, errors.ErrBadRequest
		}
		return svc.ProcessFile(ctx, req)
	}
}

func MakeHealthEndpoint(svc service.FileHandlerService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return svc.Health(ctx)
	}
}

// Endpoints save all endpoints
type Endpoints struct {
	ProcessFile endpoint.Endpoint
	Health      endpoint.Endpoint
}

// NewEndpoints generate all endpoints
func NewEndpoints(svc service.FileHandlerService) Endpoints {
	return Endpoints{
		ProcessFile: MakeProcessFileEndpoint(svc),
		Health:      MakeHealthEndpoint(svc),
	}
}
