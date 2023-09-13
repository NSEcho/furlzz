package main

import (
	_ "embed"
	"fmt"
	"github.com/nsecho/furlzz/cmd"
	"os"
)

//go:embed script/script.js
var scriptContent string

func main() {
	if err := cmd.Execute(scriptContent); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
