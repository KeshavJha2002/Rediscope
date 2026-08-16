package cli

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

type ServerOptions struct {
	OutDir      string
	Port        int
	OpenBrowser bool
}

func StartServer(opts ServerOptions) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("server internal panic: %v", r)
		}
	}()

	indexPath := filepath.Join(opts.OutDir, "index.html")
	if _, statErr := os.Stat(indexPath); statErr != nil {
		return fmt.Errorf("viewer index file not found at %s: %w", indexPath, statErr)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			http.ServeFile(w, r, indexPath)
			return
		}
		targetPath := filepath.Join(opts.OutDir, filepath.Clean(r.URL.Path))
		if info, statErr := os.Stat(targetPath); statErr == nil && !info.IsDir() {
			http.ServeFile(w, r, targetPath)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, indexPath)
	})

	var listener net.Listener

	if opts.Port > 0 {
		l, listenErr := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", opts.Port))
		if listenErr != nil {
			return fmt.Errorf("failed to bind to port %d: %w", opts.Port, listenErr)
		}
		listener = l
	} else {
		preferredPorts := []int{3000, 3001, 3002, 3847, 8080, 8081}
		for _, p := range preferredPorts {
			l, listenErr := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", p))
			if listenErr == nil {
				listener = l
				break
			}
		}
		if listener == nil {
			l, listenErr := net.Listen("tcp", "127.0.0.1:0")
			if listenErr != nil {
				return fmt.Errorf("failed to start server on ephemeral port: %w", listenErr)
			}
			listener = l
		}
	}

	actualPort := opts.Port
	if tcpAddr, ok := listener.Addr().(*net.TCPAddr); ok {
		actualPort = tcpAddr.Port
	}
	url := fmt.Sprintf("http://localhost:%d", actualPort)

	fmt.Fprintf(os.Stdout, "\n  ➜  RDB Viewer live at: %s\n", url)
	fmt.Fprintln(os.Stdout, "  ➜  Press Ctrl+C to stop.")

	if opts.OpenBrowser {
		go openBrowser(url)
	}

	server := &http.Server{
		Handler: mux,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stop
		fmt.Fprintln(os.Stdout, "\nStopping RDB Viewer server...")
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}()

	if serveErr := server.Serve(listener); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
		return serveErr
	}

	return nil
}

func openBrowser(url string) {
	defer func() {
		_ = recover() // Safe boundary: never panic in background browser opener
	}()

	if os.Getenv("CI") != "" || (os.Getenv("NO_COLOR") != "" && os.Getenv("TERM") == "dumb") {
		return
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if cmd != nil {
		_ = cmd.Start()
	}
}
