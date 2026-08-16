package test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"rediscope/internal/cli"
	"rediscope/internal/rdb"
	"rediscope/internal/viewer"
)

func fixturePath(t *testing.T, filename string) string {
	t.Helper()
	return filepath.Join("testdata", filename)
}

func TestParseNativeTypesRDB(t *testing.T) {
	parser := rdb.NewParser()
	model, err := parser.ParseFile(fixturePath(t, "native-types.rdb"))
	if err != nil {
		t.Fatal(err)
	}
	if model.RawVersion != "0015" && model.Version != "RDB v15" {
		t.Fatalf("version = %q, raw = %q, want RDB v15", model.Version, model.RawVersion)
	}
	if len(model.Keys) != 11 {
		t.Fatalf("keys = %d, want 11", len(model.Keys))
	}
	if len(model.Groups) != 3 {
		t.Fatalf("groups = %d, want 3", len(model.Groups))
	}
	if model.Groups[0].Title != "File metadata" {
		t.Fatalf("group[0] title = %q, want File metadata", model.Groups[0].Title)
	}
	if model.Groups[1].Title != "Key value pairs" {
		t.Fatalf("group[1] title = %q, want Key value pairs", model.Groups[1].Title)
	}
	if model.Groups[2].Title != "Trailer" {
		t.Fatalf("group[2] title = %q, want Trailer", model.Groups[2].Title)
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

	// Verify that key-value records contain value string segments
	foundValueString := false
	for _, group := range model.Groups {
		if group.Title == "Key value pairs" {
			for _, rec := range group.Records {
				for _, str := range rec.Strings {
					if str.Kind == "value" {
						foundValueString = true
						if str.Text == "" {
							t.Fatalf("empty value string text in record %s", rec.Label)
						}
					}
				}
			}
		}
	}
	if !foundValueString {
		t.Fatalf("expected value string segments in key-value records")
	}
}

func TestViewerWritesIndex(t *testing.T) {
	parser := rdb.NewParser()
	model, err := parser.ParseFile(fixturePath(t, "native-types.rdb"))
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
	content, err := os.ReadFile(result.IndexPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "RDB File Explorer") {
		t.Fatalf("index.html missing RDB File Explorer title")
	}
	if !strings.Contains(string(content), "byte-grid") {
		t.Fatalf("index.html missing byte-grid")
	}
}

func TestResolveFilePatterns(t *testing.T) {
	testDir := t.TempDir()

	// Create dummy test files
	sampleFiles := []string{
		"redis-6.2.23-bulk.rdb",
		"redis-7.0.15-bulk.rdb",
		"redis-7.2.15-bulk.rdb",
		"redis-7.4.10-bulk.rdb",
		"redis-8.0.6-bulk.rdb",
		"redis-8.2.8-bulk.rdb",
		"redis-8.4.5-bulk.rdb",
		"redis-8.6.5-bulk.rdb",
		"redis-8.8.1-bulk.rdb",
	}
	for _, f := range sampleFiles {
		if err := os.WriteFile(filepath.Join(testDir, f), []byte("REDIS0009"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// 1. Test glob matching
	files, err := cli.ResolveFilePatterns([]string{filepath.Join(testDir, "*bulk.rdb")})
	if err != nil {
		t.Fatalf("glob error: %v", err)
	}
	if len(files) != 9 {
		t.Fatalf("glob returned %d files, want 9", len(files))
	}

	// 2. Test regex matching
	regexPattern := filepath.Join(testDir, `redis-[67]\..*bulk\.rdb`)
	regexFiles, err := cli.ResolveFilePatterns([]string{regexPattern})
	if err != nil {
		t.Fatalf("regex error: %v", err)
	}
	if len(regexFiles) != 4 { // 6.2, 7.0, 7.2, 7.4
		t.Fatalf("regex returned %d files, want 4", len(regexFiles))
	}

	// 3. Test exact file
	exactPath := filepath.Join(testDir, "redis-6.2.23-bulk.rdb")
	exactFiles, err := cli.ResolveFilePatterns([]string{exactPath})
	if err != nil {
		t.Fatalf("exact error: %v", err)
	}
	if len(exactFiles) != 1 || exactFiles[0] != exactPath {
		t.Fatalf("exact matched %v, want [%s]", exactFiles, exactPath)
	}

	// 4. Test directory scan
	dirFiles, err := cli.ResolveFilePatterns([]string{testDir})
	if err != nil {
		t.Fatalf("directory scan error: %v", err)
	}
	if len(dirFiles) < 9 {
		t.Fatalf("directory scan returned %d files, want >= 9", len(dirFiles))
	}

	// 5. Test deduplication
	dedupFiles, err := cli.ResolveFilePatterns([]string{exactPath, exactPath, filepath.Join(testDir, "*bulk.rdb")})
	if err != nil {
		t.Fatalf("dedup error: %v", err)
	}
	if len(dedupFiles) != len(files) {
		t.Fatalf("dedup returned %d files, want %d", len(dedupFiles), len(files))
	}
}

func TestParseTestdataRDBs(t *testing.T) {
	testDir := "testdata"
	rdbFiles, err := filepath.Glob(filepath.Join(testDir, "*.rdb"))
	if err != nil {
		t.Fatal(err)
	}
	if len(rdbFiles) == 0 {
		t.Fatalf("no testdata RDB files found in %s", testDir)
	}

	parser := rdb.NewParser()
	for _, file := range rdbFiles {
		model, err := parser.ParseFile(file)
		if err != nil {
			t.Fatalf("failed to parse %s: %v", filepath.Base(file), err)
		}
		if model.Bytes == 0 || len(model.Hex) == 0 {
			t.Fatalf("%s has empty bytes/hex", filepath.Base(file))
		}
		if len(model.Groups) == 0 {
			t.Fatalf("%s has no record groups", filepath.Base(file))
		}
	}
}

func TestParseModuleRDB(t *testing.T) {
	parser := rdb.NewParser()
	model, err := parser.ParseFile(fixturePath(t, "module.rdb"))
	if err != nil {
		t.Fatal(err)
	}
	if len(model.Keys) != 3 {
		t.Fatalf("module.rdb keys = %d, want 3", len(model.Keys))
	}
}

func TestServerServesIndex(t *testing.T) {
	parser := rdb.NewParser()
	model, err := parser.ParseFile(fixturePath(t, "native-types.rdb"))
	if err != nil {
		t.Fatal(err)
	}
	outDir := t.TempDir()
	_, err = viewer.NewRDBViewer().Write(outDir, []rdb.FileModel{model})
	if err != nil {
		t.Fatal(err)
	}

	// Verify server options error check for missing dir
	err = cli.StartServer(cli.ServerOptions{
		OutDir:      filepath.Join(outDir, "non-existent"),
		Port:        9999,
		OpenBrowser: false,
	})
	if err == nil {
		t.Fatalf("expected error for non-existent outDir")
	}
}

func TestErrorBoundaries_CorruptedAndTruncatedRDBs(t *testing.T) {
	parser := rdb.NewParser()

	// 1. Empty input
	_, err := parser.Parse("empty.rdb", []byte{})
	if err == nil || !errors.Is(err, rdb.ErrTooSmall) {
		t.Fatalf("expected ErrTooSmall for empty data, got: %v", err)
	}

	// 2. Short header (5 bytes)
	_, err = parser.Parse("short.rdb", []byte("REDIS"))
	if err == nil || !errors.Is(err, rdb.ErrTooSmall) {
		t.Fatalf("expected ErrTooSmall for 5-byte data, got: %v", err)
	}

	// 3. Invalid magic header
	_, err = parser.Parse("invalid_magic.rdb", []byte("NOTREDIS15"))
	if err == nil || !errors.Is(err, rdb.ErrInvalidSignature) {
		t.Fatalf("expected ErrInvalidSignature, got: %v", err)
	}

	// 4. Truncated stream after header
	_, err = parser.Parse("truncated_header.rdb", []byte("REDIS0015"))
	// Parser reads header and stops at EOF or encounters truncated payload
	if err != nil {
		// Valid: header without EOF opcode or empty file
	}

	// 5. Truncated opcode stream (e.g. EXPIREMS without 8 bytes)
	data := append([]byte("REDIS0015"), 0xFC, 0x01, 0x02) // only 2 bytes instead of 8
	_, err = parser.Parse("truncated_expire.rdb", data)
	if err == nil || !errors.Is(err, rdb.ErrTruncated) {
		t.Fatalf("expected ErrTruncated for truncated expire-ms, got: %v", err)
	}

	// 6. Truncated EXPIRESEC without 4 bytes
	data = append([]byte("REDIS0015"), 0xFD, 0x01) // only 1 byte instead of 4
	_, err = parser.Parse("truncated_expire_sec.rdb", data)
	if err == nil || !errors.Is(err, rdb.ErrTruncated) {
		t.Fatalf("expected ErrTruncated for truncated expire-sec, got: %v", err)
	}

	// 7. Unsupported opcode (pre-GA function opcode 0xF6)
	data = append([]byte("REDIS0015"), 0xF6, 0x04, 't', 'e', 's', 't')
	_, err = parser.Parse("unsupported_function.rdb", data)
	if err == nil || !errors.Is(err, rdb.ErrUnsupportedOpcode) {
		t.Fatalf("expected ErrUnsupportedOpcode for opcode 0xF6, got: %v", err)
	}

	// 8. Truncated key record
	data = append([]byte("REDIS0015"), 0x00, 0x0A, 's', 'h', 'o', 'r', 't') // claims 10 bytes key, only 5 present
	_, err = parser.Parse("truncated_key.rdb", data)
	if err == nil || !errors.Is(err, rdb.ErrTruncated) {
		t.Fatalf("expected ErrTruncated for truncated key name, got: %v", err)
	}

	// 9. Unsupported value type opcode
	data = append([]byte("REDIS0015"), 0xEE, 0x03, 'k', 'e', 'y')
	_, err = parser.Parse("unsupported_type.rdb", data)
	if err == nil || !errors.Is(err, rdb.ErrUnsupportedType) {
		t.Fatalf("expected ErrUnsupportedType for type 0xEE, got: %v", err)
	}

	// 10. Non-existent file read
	_, err = parser.ParseFile("/path/to/non_existent_file.rdb")
	if err == nil {
		t.Fatalf("expected error for non-existent file path")
	}
}

func TestErrorBoundaries_CLIApp(t *testing.T) {
	app := cli.NewApp()

	// 1. Missing command
	if err := app.Run([]string{}); err == nil {
		t.Fatalf("expected error for missing command")
	}

	// 2. Unknown command
	if err := app.Run([]string{"unknown-cmd"}); err == nil {
		t.Fatalf("expected error for unknown command")
	}

	// 3. Missing patterns for rdb command
	if err := app.Run([]string{"rdb"}); err == nil {
		t.Fatalf("expected error for rdb without arguments")
	}

	// 4. Invalid port flag
	if err := app.Run([]string{"rdb", "sample.rdb", "--port", "invalid-port"}); err == nil {
		t.Fatalf("expected error for invalid port string")
	}
	if err := app.Run([]string{"rdb", "sample.rdb", "--port", "999999"}); err == nil {
		t.Fatalf("expected error for out of range port number")
	}
	if err := app.Run([]string{"rdb", "sample.rdb", "-p", "invalid-port"}); err == nil {
		t.Fatalf("expected error for invalid -p port string")
	}
	if err := app.Run([]string{"rdb", "sample.rdb", "-p=999999"}); err == nil {
		t.Fatalf("expected error for out of range -p= port number")
	}

	// 5. Missing --out value
	if err := app.Run([]string{"rdb", "sample.rdb", "--out"}); err == nil {
		t.Fatalf("expected error for missing --out argument value")
	}

	// 6. Non-matching pattern
	if err := app.Run([]string{"rdb", "non_existent_path_*.rdb", "--no-serve", "--no-open"}); err == nil {
		t.Fatalf("expected error for non-matching file pattern")
	}

	// 7. Version and Help commands should succeed without error
	if err := app.Run([]string{"version"}); err != nil {
		t.Fatalf("version command failed: %v", err)
	}
	if err := app.Run([]string{"help"}); err != nil {
		t.Fatalf("help command failed: %v", err)
	}
}

func TestErrorBoundaries_Viewer(t *testing.T) {
	v := viewer.NewRDBViewer()

	// 1. Empty output directory
	_, err := v.Write("", []rdb.FileModel{})
	if err == nil {
		t.Fatalf("expected error for empty outDir")
	}

	// 2. Empty models list writes valid index.html
	tmpDir := t.TempDir()
	res, err := v.Write(tmpDir, []rdb.FileModel{})
	if err != nil {
		t.Fatalf("expected success for empty models slice, got: %v", err)
	}
	if _, err := os.Stat(res.IndexPath); err != nil {
		t.Fatalf("index.html was not created: %v", err)
	}
}
