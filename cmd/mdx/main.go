package main

import (
	"fmt"
	"os"

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
	var outputPath string

	mdxCmd := &cobra.Command{
		Use:   "mdx",
		Short: "Export journals to MDX",
		Long:  `Export all journals from a world to MDX format.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMDX(worldName, outputPath, configPath)
		},
	}

	mdxCmd.Flags().StringVarP(&worldName, "world", "w", "", "World name to export (required)")
	mdxCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output directory path (required)")
	mdxCmd.Flags().StringVarP(&configPath, "config", "c", "", "Config file path")

	rootCmd.AddCommand(mdxCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runMDX(worldName, outputPath, configPath string) error {
	if worldName == "" {
		return fmt.Errorf("--world flag is required")
	}

	if outputPath == "" {
		return fmt.Errorf("--output flag is required")
	}

	// Validate world exists
	worldPath := fmt.Sprintf("worlds/%s", worldName)
	if _, err := os.Stat(worldPath); os.IsNotExist(err) {
		return fmt.Errorf("world not found: %s", worldPath)
	}

	fmt.Printf("Exporting journals from world: %s\n", worldName)
	fmt.Printf("Output directory: %s\n", outputPath)

	// TODO: Implement MDX export
	// repo := journal.NewRepository(...)
	// generator := mdx.NewGenerator(outputPath)
	// generator.Export(worldName)

	fmt.Println("MDX export completed (placeholder)")
	return nil
}
