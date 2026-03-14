package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/anomalyco/fvtt-journal-mcp/internal/journal"
	"github.com/anomalyco/fvtt-journal-mcp/internal/mcp"
	"github.com/anomalyco/fvtt-journal-mcp/internal/mcp/tools"
	"github.com/anomalyco/fvtt-journal-mcp/internal/mdx"
	"github.com/spf13/cobra"
)

var (
	// CLI flags
	WorldsPath    string
	ConfigPath    string
	mdxWorldsPath string
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
		Long:  `Start the MCP server. The server will automatically discover all worlds in the --worlds folder.`,
		RunE:  runServe,
	}

	serveCmd.Flags().StringVarP(&WorldsPath, "worlds", "w", "", "WORLDS folder path (required, e.g., ./worlds)")
	serveCmd.Flags().StringVarP(&ConfigPath, "config", "c", "", "Config file path")

	mdxCmd := &cobra.Command{
		Use:   "mdx",
		Short: "Export journals to MDX",
		Long:  `Export all journals from the --worlds folder to MDX format.`,
		RunE:  runMDX,
	}

	mdxCmd.Flags().StringVarP(&mdxWorldsPath, "worlds", "w", "", "WORLDS folder path (required, e.g., ./worlds)")
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
	if WorldsPath == "" {
		WorldsPath = "./worlds"
	}

	// Validate worlds directory exists
	if _, err := os.Stat(WorldsPath); os.IsNotExist(err) {
		return fmt.Errorf("worlds directory not found: %s", WorldsPath)
	}

	// Discover all worlds
	availableWorlds, err := journal.ListWorlds(WorldsPath)
	if err != nil {
		return fmt.Errorf("failed to list worlds: %w", err)
	}

	if len(availableWorlds) == 0 {
		return fmt.Errorf("no worlds found in %s", WorldsPath)
	}

	// Initialize logger (writes to stderr)
	logger := log.New(os.Stderr, "[FJM] ", log.LstdFlags)

	logger.Printf("Starting MCP server")
	logger.Printf("Worlds path: %s", WorldsPath)
	logger.Printf("Discovered %d worlds: %v", len(availableWorlds), availableWorlds)

	// Initialize MCP server (stdin/stdout for JSON-RPC)
	server := mcp.NewServer(os.Stdin, os.Stdout)

	// Initialize and register all tools
	registry := tools.NewRegistry(WorldsPath)
	registry.RegisterAll(server)

	logger.Printf("Registered %d MCP tools", len(registry.GetTools()))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Println("Received shutdown signal")
		cancel()
	}()

	// Start MCP server
	go func() {
		if err := server.Start(); err != nil {
			logger.Printf("MCP server error: %v", err)
			cancel()
		}
	}()

	<-ctx.Done()

	logger.Println("Server shut down")
	return nil
}

func runMDX(cmd *cobra.Command, args []string) error {
	if mdxOutputPath == "" {
		return fmt.Errorf("--output flag is required")
	}

	// Validate required flags
	if mdxWorldsPath == "" {
		mdxWorldsPath = "./worlds" // Default
	}

	// Validate worlds directory exists
	if _, err := os.Stat(mdxWorldsPath); os.IsNotExist(err) {
		return fmt.Errorf("worlds directory not found: %s", mdxWorldsPath)
	}

	// Discover all worlds
	availableWorlds, err := journal.ListWorlds(mdxWorldsPath)
	if err != nil {
		return fmt.Errorf("failed to list worlds: %w", err)
	}

	if len(availableWorlds) == 0 {
		return fmt.Errorf("no worlds found in %s", mdxWorldsPath)
	}

	log.Printf("Exporting journals from worlds in: %s", mdxWorldsPath)
	log.Printf("Discovered %d worlds: %v", len(availableWorlds), availableWorlds)
	log.Printf("Output directory: %s", mdxOutputPath)

	for _, worldName := range availableWorlds {
		log.Printf("Exporting world: %s", worldName)
		repo, err := journal.NewRepository(mdxWorldsPath, worldName)
		if err != nil {
			log.Printf("Failed to open world %s: %v", worldName, err)
			continue
		}

		generator := mdx.NewGenerator(mdxOutputPath, mdxWorldsPath, worldName)
		if err := generator.Export(repo, worldName); err != nil {
			log.Printf("Failed to export world %s: %v", worldName, err)
			repo.Close()
			continue
		}

		repo.Close()
		log.Printf("Successfully exported world: %s", worldName)
	}

	log.Printf("MDX export completed to %s", mdxOutputPath)
	return nil
}
