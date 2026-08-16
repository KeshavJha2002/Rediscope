package main

import (
	"fmt"
	"os"

	"rediscope/internal/cli"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "rediscope fatal error: %v\n", r)
			os.Exit(2)
		}
	}()

	if err := cli.NewApp().Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "rediscope: %v\n", err)
		os.Exit(1)
	}
}
