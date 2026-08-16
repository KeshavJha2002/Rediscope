package viewer

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"rediscope/internal/rdb"
)

type RDBViewer struct{}

type WriteResult struct {
	IndexPath string
}

func NewRDBViewer() *RDBViewer {
	return &RDBViewer{}
}

func (v *RDBViewer) Write(outDir string, files []rdb.FileModel) (res WriteResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("viewer runtime panic: %v", r)
		}
	}()

	if outDir == "" {
		return WriteResult{}, errors.New("output directory cannot be empty")
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return WriteResult{}, fmt.Errorf("failed to create viewer output directory %s: %w", outDir, err)
	}

	payload, err := json.Marshal(files)
	if err != nil {
		return WriteResult{}, fmt.Errorf("failed to serialize RDB models to JSON: %w", err)
	}

	indexPath := filepath.Join(outDir, "index.html")
	f, err := os.Create(indexPath)
	if err != nil {
		return WriteResult{}, fmt.Errorf("failed to create index.html at %s: %w", indexPath, err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close %s: %w", indexPath, closeErr)
		}
	}()

	tpl, err := template.New("viewer").Parse(htmlTemplate)
	if err != nil {
		return WriteResult{}, fmt.Errorf("failed to parse HTML template: %w", err)
	}

	if err := tpl.Execute(f, struct {
		Payload template.JS
	}{
		Payload: template.JS(payload),
	}); err != nil {
		return WriteResult{}, fmt.Errorf("failed to render HTML template to %s: %w", indexPath, err)
	}

	return WriteResult{IndexPath: indexPath}, nil
}
