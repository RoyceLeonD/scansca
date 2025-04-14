package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello from Scansca MCP Server!")
	fmt.Println("This is a minimal implementation to make the build process work.")
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}