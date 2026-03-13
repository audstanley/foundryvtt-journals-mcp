package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/anomalyco/fvtt-journal-mcp/internal/journal"
	"github.com/anomalyco/fvtt-journal-mcp/internal/mdx"
	"github.com/spf13/cobra"
)

var (
	// CLI flags
	WorldsPath    string
	WorldName     string
	ConfigPath    string
	mdxWorldsPath string
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
		Long:  `Start the MCP server for a specific world. Requires --worlds (WORLDS folder) and --name (world name).`,
		RunE:  runServe,
	}

	serveCmd.Flags().StringVarP(&WorldsPath, "worlds", "w", "", "WORLDS folder path (required, e.g., ./worlds)")
	serveCmd.Flags().StringVarP(&WorldName, "name", "n", "", "World name to serve (required, e.g., MyWorld)")
	serveCmd.Flags().StringVarP(&ConfigPath, "config", "c", "", "Config file path")

	mdxCmd := &cobra.Command{
		Use:   "mdx",
		Short: "Export journals to MDX",
		Long:  `Export all journals from a world to MDX format. Requires --worlds (WORLDS folder) and --name (world name).`,
		RunE:  runMDX,
	}

	mdxCmd.Flags().StringVarP(&mdxWorldsPath, "worlds", "w", "", "WORLDS folder path (required, e.g., ./worlds)")
	mdxCmd.Flags().StringVarP(&mdxWorldName, "name", "n", "", "World name to export (required, e.g., MyWorld)")
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
	// Validate required flags
	if WorldName == "" {
		return fmt.Errorf("--name flag is required")
	}
	if WorldsPath == "" {
		WorldsPath = "./worlds" // Default
	}

	// Validate world exists
	worldPath := fmt.Sprintf("%s/%s", WorldsPath, WorldName)
	if _, err := os.Stat(worldPath); os.IsNotExist(err) {
		return fmt.Errorf("world not found: %s", worldPath)
	}

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
		return fmt.Errorf("--name flag is required")
	}

	if mdxOutputPath == "" {
		return fmt.Errorf("--output flag is required")
	}

	// Validate required flags
	if mdxWorldsPath == "" {
		mdxWorldsPath = "./worlds" // Default
	}

	// Validate world exists
	worldPath := fmt.Sprintf("%s/%s", mdxWorldsPath, mdxWorldName)
	if _, err := os.Stat(worldPath); os.IsNotExist(err) {
		return fmt.Errorf("world not found: %s", worldPath)
	}

	log.Printf("Exporting journals from world: %s", mdxWorldName)
	log.Printf("WORLDS path: %s", mdxWorldsPath)
	log.Printf("World path: %s", worldPath)
	log.Printf("Output directory: %s", mdxOutputPath)

	repo, err := journal.NewRepository(mdxWorldsPath, mdxWorldName)
	if err != nil {
		return fmt.Errorf("failed to open world: %w", err)
	}
	defer repo.Close()

	generator := mdx.NewGenerator(mdxOutputPath)
	if err := generator.Export(repo, mdxWorldName); err != nil {
		return fmt.Errorf("failed to export: %w", err)
	}

	log.Printf("MDX export completed to %s", mdxOutputPath)
	return nil
}
