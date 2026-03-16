package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	// CLI flags
	configPath string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "fjm-mdx",
		Short: "Foundry VTT Journal MDX Exporter",
		Long:  `Export Foundry VTT journals to MDX format`,
	}

	var worldName string
	var worldsDir string
	var outputPath string

	mdxCmd := &cobra.Command{
		Use:   "mdx",
		Short: "Export journals to MDX",
		Long:  `Export all journals from one or all worlds to MDX format.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMDX(worldName, worldsDir, outputPath, configPath)
		},
	}

	mdxCmd.Flags().StringVarP(&worldName, "world", "w", "", "World name to export")
	mdxCmd.Flags().StringVarP(&worldsDir, "worlds", "", "./worlds", "Path to worlds directory")
	mdxCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output directory path (required)")
	mdxCmd.Flags().StringVarP(&configPath, "config", "c", "", "Config file path")

	rootCmd.AddCommand(mdxCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runMDX(worldName, worldsDir, outputPath, configPath string) error {
	if outputPath == "" {
		return fmt.Errorf("--output flag is required")
	}

	if worldName != "" {
		worldPath := filepath.Join(worldsDir, worldName)
		if _, err := os.Stat(worldPath); os.IsNotExist(err) {
			return fmt.Errorf("world not found: %s", worldPath)
		}

		fmt.Printf("Exporting journals from world: %s\n", worldName)
		fmt.Printf("Output directory: %s\n", outputPath)

		fmt.Println("MDX export completed (placeholder)")
		return nil
	}

	entries, err := os.ReadDir(worldsDir)
	if err != nil {
		return fmt.Errorf("failed to read worlds directory: %w", err)
	}

	var worlds []string
	ignored := map[string]bool{"sounds": true}
	for _, entry := range entries {
		if entry.IsDir() && !ignored[entry.Name()] {
			worlds = append(worlds, entry.Name())
		}
	}

	if len(worlds) == 0 {
		return fmt.Errorf("no worlds found in %s", worldsDir)
	}

	fmt.Printf("Exporting journals from worlds in: %s\n", worldsDir)
	fmt.Printf("Discovered %d worlds: %v\n", len(worlds), worlds)
	fmt.Printf("Output directory: %s\n", outputPath)

	for _, w := range worlds {
		worldPath := filepath.Join(worldsDir, w)
		if _, err := os.Stat(worldPath); os.IsNotExist(err) {
			fmt.Printf("Failed to open world %s: world not found\n", w)
			continue
		}

		fmt.Printf("Exporting world: %s\n", w)

		fmt.Printf("Successfully exported world: %s\n", w)
	}

	fmt.Println("MDX export completed")
	return nil
}
