package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/anomalyco/fvtt-journal-mcp/pkg/config"
	"github.com/spf13/cobra"
)

var (
	// CLI flags
	WorldName     string
	ConfigPath    string
	mdxWorldName  string
	mdxOutputPath string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "fjm",
		Short: "Foundry VTT Journal MCP Server",
		Long:  "An MCP server for reading and searching Foundry VTT journals",
	}

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the MCP server",
		Long:  `Start the MCP server for a specific world. The server requires a --world parameter.`,
		RunE:  runServe,
	}

	serveCmd.Flags().StringVarP(&WorldName, "world", "w", "", "World name to serve (required)")
	serveCmd.Flags().StringVarP(&ConfigPath, "config", "c", "", "Config file path")

	mdxCmd := &cobra.Command{
		Use:   "mdx",
		Short: "Export journals to MDX",
		Long:  `Export all journals from a world to MDX format.`,
		RunE:  runMDX,
	}

	var mdxWorldName string
	var mdxOutputPath string
	mdxCmd.Flags().StringVarP(&mdxWorldName, "world", "w", "", "World name to export (required)")
	mdxCmd.Flags().StringVarP(&mdxOutputPath, "output", "o", "", "Output directory path (required)")
	mdxCmd.Flags().StringVarP(&ConfigPath, "config", "c", "", "Config file path")

	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(mdxCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runServe(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load(ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override world name from CLI
	if WorldName != "" {
		// World is specified via CLI flag
	} else if cfg.User != "" {
		WorldName = cfg.User // Fallback if needed
	}

	if WorldName == "" {
		return fmt.Errorf("--world flag is required")
	}

	// Validate world exists
	worldPath := fmt.Sprintf("worlds/%s", WorldName)
	if _, err := os.Stat(worldPath); os.IsNotExist(err) {
		return fmt.Errorf("world not found: %s", worldPath)
	}

	log.Printf("Starting MCP server for world: %s", WorldName)
	log.Printf("Worlds path: %s", cfg.WorldsPath)
	log.Printf("Username for permissions: %s", cfg.User)

	// TODO: Implement MCP server initialization
	// server := mcp.NewServer(...)
	// server.Run()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal")
		cancel()
	}()

	// Server loop would go here
	<-ctx.Done()

	log.Println("Server shut down")
	return nil
}

func runMDX(cmd *cobra.Command, args []string) error {
	if mdxWorldName == "" {
		return fmt.Errorf("--world flag is required")
	}

	if mdxOutputPath == "" {
		return fmt.Errorf("--output flag is required")
	}

	// Load config
	cfg, err := config.Load(ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate world exists
	worldPath := fmt.Sprintf("worlds/%s", mdxWorldName)
	if _, err := os.Stat(worldPath); os.IsNotExist(err) {
		return fmt.Errorf("world not found: %s", worldPath)
	}

	log.Printf("Exporting journals from world: %s", mdxWorldName)
	log.Printf("Output directory: %s", mdxOutputPath)
	log.Printf("Worlds path: %s", cfg.WorldsPath)

	// TODO: Implement MDX export
	// repo := journal.NewRepository(...)
	// generator := mdx.NewGenerator(mdxOutputPath)
	// generator.Export(mdxWorldName)

	log.Println("MDX export completed (placeholder)")
	return nil
}
