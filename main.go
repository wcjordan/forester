package main

import (
	"fmt"
	"os"

	"forester/game"
)

func main() {
	g := game.New()
	if err := g.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
