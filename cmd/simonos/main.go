package main

import (
	"fmt"
	"os"

	"simonos/internal/config"
	"simonos/internal/workspace"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := newRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "simonos",
		Short: "SimonOS local-first agent runtime",
	}

	cmd.PersistentFlags().StringVar(&configPath, "config", "configs/config.example.yaml", "Path to config file")
	cmd.AddCommand(newRunCmd(&configPath))
	cmd.AddCommand(newChatCmd(&configPath))
	cmd.AddCommand(newConfigCmd(&configPath))

	return cmd
}

func loadConfig(path string) (*config.Config, error) {
	cfg, err := config.Load(path)
	if err != nil {
		return nil, err
	}

	if cfg.Workspace.Root == "" {
		cfg.Workspace.Root = "."
	}

	ws, err := workspace.New(cfg.Workspace.Root)
	if err != nil {
		return nil, err
	}
	cfg.Workspace.Root = ws.RootPath

	return cfg, nil
}
