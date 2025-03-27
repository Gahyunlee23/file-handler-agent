package agent

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Supported input file formats
var supportedInputFormats = map[string]bool{
	".pdf": true,
}

// Supported output image formats
var supportedOutputFormats = map[string]bool{
	"jpg":  true,
	"jpeg": true,
	"png":  true,
}

// GhostscriptAgent execute ghostscript cmd
type GhostscriptAgent struct {
	BinaryPath string
	OutputDir  string
}

// NewGhostscriptAgent generate new ghostscript agent
func NewGhostscriptAgent(binaryPath, outputDir string) *GhostscriptAgent {
	return &GhostscriptAgent{
		BinaryPath: binaryPath,
		OutputDir:  outputDir,
	}
}

// Execute new ghostscript commands
func (g *GhostscriptAgent) Execute(ctx context.Context, action string, params map[string]interface{}, files []string) ([]string, error) {
	switch action {
	case "convertPdfToImage":
		return g.convertPdfToImage(ctx, params, files)
	default:
		return nil, errors.New("unsupported action")
	}
}

// convertPdfToImage converting image extension
func (g *GhostscriptAgent) convertPdfToImage(ctx context.Context, params map[string]interface{}, files []string) ([]string, error) {
	// Check for context cancellation early
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// Continue with processing
	}

	startTime := time.Now()
	log.Printf("Starting PDF to image conversion for %d files", len(files))

	if len(files) == 0 {
		return nil, errors.New("no input files provided")
	}

	// Validate input files format
	for _, inputFile := range files {
		if !fileExists(inputFile) {
			return nil, fmt.Errorf("input file not found: %s", inputFile)
		}

		ext := filepath.Ext(inputFile)
		if !supportedInputFormats[ext] {
			return nil, fmt.Errorf("unsupported input file format: %s", ext)
		}
	}

	// Parse and validate parameters
	resolution, _ := params["resolution"].(float64)
	if resolution == 0 {
		resolution = 72
	}

	imageFormat, _ := params["image_format"].(string)
	if imageFormat == "" {
		imageFormat = "png"
	}

	// Check if output format is supported
	if !supportedOutputFormats[imageFormat] {
		return nil, fmt.Errorf("unsupported output image format: %s", imageFormat)
	}

	antiAliasing, _ := params["anti_aliasing"].(bool)

	// If pdf is multiple pages, select the page of user's choice
	pages, _ := params["pages"].(string)
	if pages == "" {
		pages = "1"
	}

	outputDir, _ := params["output_dir"].(string)
	if outputDir == "" {
		outputDir = "output"
	}

	// Create output directory only after all validations passed
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %v", err)
	}

	var outputFiles []string
	var mu sync.Mutex // Mutex to protect the outputFiles slice
	var wg sync.WaitGroup
	errCh := make(chan error, len(files)) // Channel to collect errors

	for i, inputFile := range files {
		fileStartTime := time.Now()
		log.Printf("[%d/%d] Processing file: %s", i+1, len(files), inputFile)

		wg.Add(1)
		go func(file string, fileIdx int) {
			defer wg.Done()

			baseName := filepath.Base(file)
			baseNameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))
			fileOutputDir := filepath.Join(outputDir, baseNameWithoutExt)

			// Create directory for this specific file
			if err := os.MkdirAll(fileOutputDir, 0755); err != nil {
				errCh <- fmt.Errorf("failed to create output directory for %s: %v", file, err)
				return
			}

			outputPattern := filepath.Join(fileOutputDir, fmt.Sprintf("%s-%%d.%s", baseNameWithoutExt, imageFormat))
			log.Printf("Output pattern for file %s: %s", file, outputPattern)

			// Process with ghostscript
			args := []string{
				"-dNOPAUSE",
				"-dBATCH",
				"-dSAFER",
				fmt.Sprintf("-r%d", int(resolution)),
				fmt.Sprintf("-sDEVICE=%s", getGsDevice(imageFormat)),
			}

			if antiAliasing {
				args = append(args, "-dTextAlphaBits=4", "-dGraphicsAlphaBits=4")
			}

			// Set page range
			if pages != "all" {
				args = append(args, fmt.Sprintf("-dFirstPage=%s", pages))
				args = append(args, fmt.Sprintf("-dLastPage=%s", pages))
			}

			args = append(args,
				fmt.Sprintf("-sOutputFile=%s", outputPattern),
				file,
			)

			// Set up command with proper context
			cmd := exec.CommandContext(ctx, g.BinaryPath, args...)

			// Set up a proper cancellation mechanism
			doneCh := make(chan struct{})
			var output []byte
			var cmdErr error

			go func() {
				output, cmdErr = cmd.CombinedOutput()
				close(doneCh)
			}()

			select {
			case <-ctx.Done():
				// Context was cancelled, try to kill the process
				if cmd.Process != nil {
					cmd.Process.Kill()
				}
				errCh <- ctx.Err()
				return
			case <-doneCh:
				// Command completed
				if cmdErr != nil {
					errCh <- fmt.Errorf("ghostscript error: %v, output: %s", cmdErr, string(output))
					return
				}
			}

			// Get generated files
			pattern := strings.Replace(outputPattern, "%d", "*", 1)
			matches, err := filepath.Glob(pattern)
			if err != nil {
				errCh <- err
				return
			}

			// Thread-safe append to outputFiles
			mu.Lock()
			outputFiles = append(outputFiles, matches...)
			mu.Unlock()

			log.Printf("[%d/%d] Completed processing file: %s (took %v)",
				fileIdx+1, len(files), file, time.Since(fileStartTime))
		}(inputFile, i)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errCh)

	// Check if any errors occurred
	for err := range errCh {
		if err != nil {
			return nil, err // Return the first error encountered
		}
	}

	totalDuration := time.Since(startTime)
	log.Printf("Completed PDF to image conversion for %d files in %v", len(files), totalDuration)

	// Add metrics for monitoring if needed
	if metricsCollector, ok := params["metrics"].(MetricsCollector); ok {
		metricsCollector.RecordDuration("pdf_conversion_time", totalDuration)
		metricsCollector.RecordCount("files_processed", len(files))
		metricsCollector.RecordCount("output_files_generated", len(outputFiles))
	}

	return outputFiles, nil
}

// getGsDevice retrieve matching ghostscript device
func getGsDevice(format string) string {
	switch format {
	case "jpg", "jpeg":
		return "jpeg"
	case "png":
		return "png16m"
	default:
		return "png16m" // Default to png if format not recognized
	}
}

// fileExists check if the file exist
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// MetricsCollector Optional metrics interface
type MetricsCollector interface {
	RecordDuration(name string, duration time.Duration)
	RecordCount(name string, count int)
}
