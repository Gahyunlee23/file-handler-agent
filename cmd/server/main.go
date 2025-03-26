package main

import (
	"file-handler-agent/pkg/endpoint"
	"file-handler-agent/pkg/service"
	"file-handler-agent/pkg/service/agent"
	httphandler "file-handler-agent/pkg/transport/http"
	"log"
	"net/http"
	"os"
)

func main() {
	ghostscriptPath := "gs"
	outputDir := "temp/output"

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err := os.MkdirAll(outputDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create output directory: %v", err)
		}
	}

	registry := agent.NewRegistry()

	gsAgent := agent.NewGhostscriptAgent(ghostscriptPath, outputDir)
	registry.Register("ghostscript", gsAgent)

	svc := service.NewFileHandlerService(registry)

	endpoints := endpoint.NewEndpoints(svc)

	handler := httphandler.NewHTTPHandler(endpoints)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on : %s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
