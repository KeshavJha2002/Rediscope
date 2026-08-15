package test

import (
	"os"
	"path/filepath"
	"testing"

	"rediscope/internal/rdb"
	"rediscope/internal/viewer"
)

func fixturePath(t *testing.T, parts ...string) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(append([]string{root}, parts...)...)
}

func TestParseNativeTypesRDB(t *testing.T) {
	parser := rdb.NewParser()
	model, err := parser.ParseFile(fixturePath(t, "lab_artifacts", "redis_persistence", "native-types.rdb"))
	if err != nil {
		t.Fatal(err)
	}
	if model.Version != "0015" {
		t.Fatalf("version = %q, want 0015", model.Version)
	}
	if len(model.Keys) != 11 {
		t.Fatalf("keys = %d, want 11", len(model.Keys))
	}
	foundHash := false
	for _, key := range model.Keys {
		if key.Key == "lab:hash" {
			foundHash = true
			if key.TypeName != "RDB_TYPE_HASH_LISTPACK" {
				t.Fatalf("lab:hash type = %s", key.TypeName)
			}
			if key.ValueEnd <= key.ValueStart {
				t.Fatalf("lab:hash value range is empty")
			}
		}
	}
	if !foundHash {
		t.Fatalf("lab:hash was not parsed")
	}
}

func TestViewerWritesIndex(t *testing.T) {
	parser := rdb.NewParser()
	model, err := parser.ParseFile(fixturePath(t, "lab_artifacts", "redis_persistence", "native-types.rdb"))
	if err != nil {
		t.Fatal(err)
	}
	outDir := t.TempDir()
	result, err := viewer.NewRDBViewer().Write(outDir, []rdb.FileModel{model, model})
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(result.IndexPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Fatalf("index.html is empty")
	}
}
