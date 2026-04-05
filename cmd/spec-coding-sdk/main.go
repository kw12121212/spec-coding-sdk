// Package main is the entry point for spec-coding-sdk.
package main

import (
	"fmt"
	"os"
)

// Version is set at build time via -ldflags.
var Version = "dev"

func main() {
	fmt.Printf("spec-coding-sdk %s\n", Version)
	os.Exit(0)
}
