package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"rediscope/internal/rdb"
	"rediscope/internal/viewer"
)

const Version = "1.0.0-beta.0"

type App struct{}

func NewApp() *App {
	return &App{}
}

func (a *App) Run(args []string) error {
	if len(args) == 0 {
		a.usage()
		return errors.New("missing command")
	}

	switch args[0] {
	case "rdb":
		return a.runRDB(args[1:])
	case "version", "-v", "--version":
		fmt.Fprintf(os.Stdout, "rediscope %s\n", Version)
		return nil
	case "help", "-h", "--help":
		a.usage()
		return nil
	default:
		a.usage()
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func (a *App) runRDB(args []string) error {
	outDir := filepath.Join(".rediscope", "rdb-viewer")
	var files []string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "help", "-h", "--help":
			a.usage()
			return nil
		case "--out":
			if i+1 >= len(args) {
				return errors.New("--out requires a directory")
			}
			outDir = args[i+1]
			i++
		default:
			files = append(files, args[i])
		}
	}

	if len(files) == 0 {
		return errors.New("usage: rediscope rdb dump.rdb [more.rdb ...] [--out dir]")
	}

	parser := rdb.NewParser()
	models := make([]rdb.FileModel, 0, len(files))
	for _, file := range files {
		model, err := parser.ParseFile(file)
		if err != nil {
			return err
		}
		models = append(models, model)
	}

	result, err := viewer.NewRDBViewer().Write(outDir, models)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Wrote RDB viewer: %s\n", result.IndexPath)
	fmt.Fprintf(os.Stdout, "Files: %d\n", len(models))
	return nil
}

func (a *App) usage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  rediscope rdb dump.rdb [more.rdb ...] [--out dir]")
	fmt.Fprintln(os.Stderr, "  rediscope version")
}
