package agent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

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
	if len(files) == 0 {
		return nil, errors.New("no input files provided")
	}

	resolution, _ := params["resolution"].(float64)
	if resolution == 0 {
		resolution = 72
	}

	imageFormat, _ := params["image_format"].(string)
	if imageFormat == "" {
		imageFormat = "png"
	}

	antiAliasing, _ := params["anti_aliasing"].(bool)

	// if pdf is multiple pages, select the page of user's choice
	pages, _ := params["pages"].(string)
	if pages == "" {
		pages = "1"
	}

	var outputFiles []string

	for _, inputFile := range files {
		if !fileExists(inputFile) {
			return nil, fmt.Errorf("input file not found: %s", inputFile)
		}

		baseName := filepath.Base(inputFile)
		baseNameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))
		outputPattern := filepath.Join(g.OutputDir, fmt.Sprintf("%s-%%d.%s", baseNameWithoutExt, imageFormat))

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

		// set page range
		if pages != "all" {
			args = append(args, fmt.Sprintf("-dFirstPage=%s", pages))
			args = append(args, fmt.Sprintf("-dLastPage=%s", pages))
		}

		args = append(args,
			fmt.Sprintf("-sOutputFile=%s", outputPattern),
			inputFile,
		)

		cmd := exec.CommandContext(ctx, g.BinaryPath, args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("ghostscript error: %v, output: %s", err, string(output))
		}

		// get generated directory
		pattern := strings.Replace(outputPattern, "%d", "*", 1)
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}

		outputFiles = append(outputFiles, matches...)
	}

	return outputFiles, nil
}

// getGsDevice retrieve matching ghostscript agent
func getGsDevice(format string) string {
	switch format {
	case "jpg", "jpeg":
		return "jpeg"
	case "png":
		return "png16m"
	case "webp":
		return "webp" // need to check if ghostscript support webp
	default:
		return "png16m"
	}
}

// fileExists check if the file exist
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
