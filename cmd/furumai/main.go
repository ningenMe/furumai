package main

import (
	"fmt"
	"io"
	"os"
)

const version = "0.1.0"

type command struct {
	name  string
	usage string
	run   func(w io.Writer, args []string) int
}

var commands = []command{
	{name: "version", usage: "Print the furumai version", run: runVersion},
	{name: "help", usage: "Show this help message", run: runHelp},
}

func main() {
	os.Exit(run(os.Stdout, os.Stderr, os.Args[1:]))
}

func run(stdout, stderr io.Writer, args []string) int {
	if len(args) == 0 {
		printUsage(stdout)
		return 1
	}

	name := args[0]
	if name == "--help" || name == "-h" {
		name = "help"
	}

	for _, cmd := range commands {
		if cmd.name == name {
			return cmd.run(stdout, args[1:])
		}
	}

	fmt.Fprintf(stderr, "unknown command: %s\n\n", name)
	printUsage(stdout)
	return 1
}

func runVersion(w io.Writer, _ []string) int {
	fmt.Fprintln(w, "furumai "+version)
	return 0
}

func runHelp(w io.Writer, _ []string) int {
	printUsage(w)
	return 0
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: furumai <command>")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	for _, cmd := range commands {
		fmt.Fprintf(w, "  %-10s %s\n", cmd.name, cmd.usage)
	}
}
