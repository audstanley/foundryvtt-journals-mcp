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

// Install builds and installs the binary to ~/go/bin
func Install() error {
	fmt.Println("Building and installing...")

	if err := Build(); err != nil {
		return err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	targetPath := homeDir + "/go/bin/fjm"

	if err := exec.Command("cp", "bin/fjm", targetPath).Run(); err != nil {
		return fmt.Errorf("failed to copy binary to %s: %w", targetPath, err)
	}

	fmt.Printf("✓ Installed fjm to %s\n", targetPath)
	return nil
}
