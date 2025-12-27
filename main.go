package main

import (
	"fmt"
	"os"

	"lazytime/cli"
	"lazytime/tui"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: lazytime <command> [args...]\n")
		fmt.Fprintf(os.Stderr, "Commands: start, stop, add, status, report, tui\n")
		os.Exit(1)
	}

	command := args[0]

	// Handle TUI separately to avoid importing tui in cli package
	if command == "tui" {
		if err := tui.LaunchTUI(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Handle all other commands through CLI
	if err := cli.RunCLI(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

