package viewer

import (
	"encoding/json"
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

func (v *RDBViewer) Write(outDir string, files []rdb.FileModel) (WriteResult, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return WriteResult{}, err
	}
	payload, err := json.Marshal(files)
	if err != nil {
		return WriteResult{}, err
	}
	indexPath := filepath.Join(outDir, "index.html")
	f, err := os.Create(indexPath)
	if err != nil {
		return WriteResult{}, err
	}
	defer f.Close()

	tpl := template.Must(template.New("viewer").Parse(htmlTemplate))
	if err := tpl.Execute(f, struct {
		Payload template.JS
	}{
		Payload: template.JS(payload),
	}); err != nil {
		return WriteResult{}, err
	}
	return WriteResult{IndexPath: indexPath}, nil
}
