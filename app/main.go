package main

import (
	"fmt"
	"os"

	"meme-factory/internal/tui"
)

func main() {
	if err := tui.Run(); err != nil {
		fmt.Printf("ERR: %v", err)
		os.Exit(1)
	}
}
