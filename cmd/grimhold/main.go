package main

import (
	"fmt"
	"os"
	"github.com/Roland-HUON/grimhold/internal/repl"
)

func main() {
	if err := repl.Start(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}