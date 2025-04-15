package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	fmt.Println("Starting Scansca MCP Server...")
	
	// Execute the actual server binary
	cmd := exec.Command("go", "run", "cmd/main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	// Forward signals to the child process
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
	
	// Wait for the process to complete
	err = cmd.Wait()
	
	// Handle exit status
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		}
		os.Exit(1)
	}
}
