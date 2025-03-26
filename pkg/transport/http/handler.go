package http

import (
	"context"
	"encoding/json"
	"file-handler-agent/pkg/endpoint"
	"file-handler-agent/pkg/service"
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// NewHTTPHandler returns HTTP request handler
func NewHTTPHandler(endpoints endpoint.Endpoints) http.Handler {
	router := mux.NewRouter()

	// Process file endpoint
	router.Methods("POST").Path("/process").Handler(httptransport.NewServer(
		endpoints.ProcessFile,
		decodeProcessFileRequest,
		encodeResponse,
	))

	router.Methods("GET").Path("/health").Handler(httptransport.NewServer(
		endpoints.Health,
		decodeHealthRequest,
		encodeResponse,
	))

	return router
}

// decodeProcessFileRequest convert from HTTP request to FileRequest
func decodeProcessFileRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req service.FileRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// decodeHealthRequest decode health request
func decodeHealthRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

// encodeResponse JSON encoding the response
func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
