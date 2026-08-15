package main

import (
	"fmt"
	"os"

	"rediscope/internal/cli"
)

func main() {
	if err := cli.NewApp().Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "rediscope: %v\n", err)
		os.Exit(1)
	}
}
