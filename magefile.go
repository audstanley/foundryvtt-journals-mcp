//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
)

// Build builds all binaries
func Build() error {
	fmt.Println("Building binary...")

	// Build single binary with both serve and mdx commands
	if err := exec.Command("go", "build", "-o", "bin/fjm", "./cmd/server").Run(); err != nil {
		return fmt.Errorf("failed to build: %w", err)
	}
	fmt.Println("✓ Built fjm")

	fmt.Println("✓ Binary built successfully")
	return nil
}

// Test runs tests
func Test() error {
	fmt.Println("Running tests...")
	return exec.Command("go", "test", "./...").Run()
}

// Clean removes build artifacts
func Clean() error {
	fmt.Println("Cleaning...")
	return os.RemoveAll("bin")
}

// Lint runs linting
func Lint() error {
	fmt.Println("Running lint...")
	if err := exec.Command("go", "fmt", "./...").Run(); err != nil {
		return err
	}
	return nil
}
