package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"rediscope/internal/rdb"
	"rediscope/internal/viewer"
)

const Version = "1.0.0"

type App struct{}

func NewApp() *App {
	return &App{}
}

func (a *App) Run(args []string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("rediscope internal CLI error: %v", r)
		}
	}()

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

func (a *App) runRDB(args []string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("rediscope RDB command internal error: %v", r)
		}
	}()

	outDir := filepath.Join(".rediscope", "rdb-viewer")
	port := 0
	openBrowser := true
	serve := true
	var rawPatterns []string

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
		case "--port", "-p":
			if i+1 >= len(args) {
				return errors.New("--port requires a port number")
			}
			p, err := strconv.Atoi(args[i+1])
			if err != nil || p < 1 || p > 65535 {
				return fmt.Errorf("invalid port number %q", args[i+1])
			}
			port = p
			i++
		case "--no-open", "--open=false":
			openBrowser = false
		case "--no-serve", "--serve=false":
			serve = false
		default:
			if strings.HasPrefix(args[i], "--out=") {
				outDir = strings.TrimPrefix(args[i], "--out=")
			} else if strings.HasPrefix(args[i], "--port=") || strings.HasPrefix(args[i], "-p=") {
				pStr := strings.TrimPrefix(strings.TrimPrefix(args[i], "--port="), "-p=")
				p, err := strconv.Atoi(pStr)
				if err != nil || p < 1 || p > 65535 {
					return fmt.Errorf("invalid port number %q", pStr)
				}
				port = p
			} else {
				rawPatterns = append(rawPatterns, args[i])
			}
		}
	}

	if port == 0 {
		if envPort := os.Getenv("PORT"); envPort != "" {
			if p, err := strconv.Atoi(envPort); err == nil && p >= 1 && p <= 65535 {
				port = p
			}
		}
	}

	if len(rawPatterns) == 0 {
		return errors.New("usage: rediscope rdb <file-or-pattern> [more-files-or-patterns ...] [--out dir] [--port port] [--no-open] [--no-serve]")
	}

	files, err := ResolveFilePatterns(rawPatterns)
	if err != nil {
		return err
	}

	parser := rdb.NewParser()
	models := make([]rdb.FileModel, 0, len(files))
	for _, file := range files {
		model, err := parser.ParseFile(file)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", file, err)
		}
		models = append(models, model)
	}

	result, err := viewer.NewRDBViewer().Write(outDir, models)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Wrote RDB viewer: %s\n", result.IndexPath)
	fmt.Fprintf(os.Stdout, "Files: %d\n", len(models))

	if serve {
		return StartServer(ServerOptions{
			OutDir:      outDir,
			Port:        port,
			OpenBrowser: openBrowser,
		})
	}

	return nil
}

func ResolveFilePatterns(patterns []string) (resolved []string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("pattern resolution internal error: %v", r)
		}
	}()

	seen := make(map[string]bool)

	addFile := func(path string) {
		clean := filepath.Clean(path)
		abs, err := filepath.Abs(clean)
		if err != nil {
			abs = clean
		}
		if !seen[abs] {
			seen[abs] = true
			resolved = append(resolved, clean)
		}
	}

	for _, pattern := range patterns {
		matchedForPattern := 0

		// 1. Direct file check
		if info, err := os.Stat(pattern); err == nil {
			if !info.IsDir() {
				addFile(pattern)
				continue
			}
			// If it's a directory, scan for .rdb files inside it
			entries, err := os.ReadDir(pattern)
			if err == nil {
				for _, entry := range entries {
					if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".rdb") {
						addFile(filepath.Join(pattern, entry.Name()))
						matchedForPattern++
					}
				}
			}
			if matchedForPattern > 0 {
				continue
			}
		}

		// 2. Glob expansion (if no ** in pattern)
		if !strings.Contains(pattern, "**") {
			matches, err := filepath.Glob(pattern)
			if err == nil && len(matches) > 0 {
				for _, m := range matches {
					if info, err := os.Stat(m); err == nil && !info.IsDir() {
						addFile(m)
						matchedForPattern++
					}
				}
			}
			if matchedForPattern > 0 {
				continue
			}
		}

		// 3. Regex / Directory pattern matching
		dir := filepath.Dir(pattern)
		base := filepath.Base(pattern)

		searchDir := dir
		if _, err := os.Stat(searchDir); err != nil {
			searchDir = "."
			base = pattern
		}

		// Try compiling directly as regex or convert glob wildcards
		var re *regexp.Regexp
		var reErr error

		if strings.Contains(base, `\`) || strings.Contains(base, "[") || strings.Contains(base, "(") {
			cleanRegex := strings.TrimPrefix(strings.TrimSuffix(base, "$"), "^")
			re, reErr = regexp.Compile("^" + cleanRegex + "$")
		}
		if re == nil || reErr != nil {
			escaped := regexp.QuoteMeta(base)
			escaped = strings.ReplaceAll(escaped, `\*`, ".*")
			escaped = strings.ReplaceAll(escaped, `\?`, ".")
			re, reErr = regexp.Compile("^" + escaped + "$")
		}

		if reErr == nil {
			entries, readErr := os.ReadDir(searchDir)
			if readErr == nil {
				for _, entry := range entries {
					if !entry.IsDir() && re.MatchString(entry.Name()) {
						addFile(filepath.Join(searchDir, entry.Name()))
						matchedForPattern++
					}
				}
			}
		}

		// 4. Recursive walk matching (for ** patterns or nested regex matches)
		if matchedForPattern == 0 && strings.Contains(pattern, "**") {
			walkRoot := searchDir
			if _, err := os.Stat(walkRoot); err != nil {
				walkRoot = "."
			}

			_ = filepath.Walk(walkRoot, func(p string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() {
					name := info.Name()
					if re != nil && (re.MatchString(name) || re.MatchString(p)) {
						addFile(p)
						matchedForPattern++
					}
				}
				return nil
			})
		}
	}

	if len(resolved) == 0 {
		return nil, fmt.Errorf("no RDB files matched given pattern(s): %s", strings.Join(patterns, ", "))
	}

	sort.Strings(resolved)
	return resolved, nil
}

func (a *App) usage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  rediscope rdb <file-or-pattern> [more-patterns ...] [--out dir] [--port port] [--no-open] [--no-serve]")
	fmt.Fprintln(os.Stderr, "  rediscope version")
}
