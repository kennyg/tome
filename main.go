package main

import (
	"os"

	"github.com/kennyg/tome/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
