package main

import (
	"context"
	"fmt"
	"os"

	"simonos/internal/agent"
	"simonos/internal/config"
	"simonos/internal/events"
	"simonos/internal/guardrails"
	"simonos/internal/memory"
	"simonos/internal/model"
	"simonos/internal/tools"
	"simonos/internal/tools/builtins"
	"simonos/internal/workspace"

	"github.com/spf13/cobra"
)

func newRunCmd(configPath *string) *cobra.Command {
	var input string

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Execute a single agent task",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, runtime, err := buildRuntime(*configPath)
			if err != nil {
				return err
			}
			defer runtime.Close()

			result, err := runtime.Agent.Run(cmd.Context(), input)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), result)
			_ = cfg
			return nil
		},
	}

	cmd.Flags().StringVar(&input, "input", "", "Input prompt to execute")
	_ = cmd.MarkFlagRequired("input")
	return cmd
}

func newChatCmd(configPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chat",
		Short: "Start a simple interactive chat loop",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, runtime, err := buildRuntime(*configPath)
			if err != nil {
				return err
			}
			defer runtime.Close()

			fmt.Fprintln(cmd.OutOrStdout(), "SimonOS chat. Press Ctrl+D to exit.")
			buffer := make([]byte, 4096)
			for {
				fmt.Fprint(cmd.OutOrStdout(), "> ")
				n, readErr := os.Stdin.Read(buffer)
				if readErr != nil {
					if n == 0 {
						fmt.Fprintln(cmd.OutOrStdout())
						return nil
					}
				}
				if n == 0 {
					return nil
				}

				result, err := runtime.Agent.Run(cmd.Context(), string(buffer[:n]))
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), result)
			}
		},
	}

	return cmd
}

func newConfigCmd(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Print the resolved configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(*configPath)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "model.default=%s\n", cfg.Model.Default)
			fmt.Fprintf(cmd.OutOrStdout(), "model.fallback=%s\n", cfg.Model.Fallback)
			fmt.Fprintf(cmd.OutOrStdout(), "memory.type=%s\n", cfg.Memory.Type)
			fmt.Fprintf(cmd.OutOrStdout(), "memory.path=%s\n", cfg.Memory.Path)
			fmt.Fprintf(cmd.OutOrStdout(), "security.require_approval=%t\n", cfg.Security.RequireApproval)
			fmt.Fprintf(cmd.OutOrStdout(), "security.allow_shell=%t\n", cfg.Security.AllowShell)
			fmt.Fprintf(cmd.OutOrStdout(), "workspace.root=%s\n", cfg.Workspace.Root)
			return nil
		},
	}
}

type runtimeDeps struct {
	Agent agent.Agent
	Close func() error
}

func buildRuntime(configPath string) (*config.Config, runtimeDeps, error) {
	cfg, err := loadConfig(configPath)
	if err != nil {
		return nil, runtimeDeps{}, err
	}

	ws, err := workspace.New(cfg.Workspace.Root)
	if err != nil {
		return nil, runtimeDeps{}, err
	}

	bus := events.NewBus(32)
	registry := tools.NewRegistry()
	registry.Register(builtins.NewFileReadTool(ws))
	registry.Register(builtins.NewFileWriteTool(ws))
	registry.Register(builtins.NewShellTool(ws, cfg.Security.AllowShell))
	executor := tools.NewExecutor(registry)

	shortTerm := memory.NewShortTermStore()
	longTerm, err := memory.NewSQLiteStore(cfg.Memory.Path)
	if err != nil {
		return nil, runtimeDeps{}, err
	}
	combined := memory.NewCombinedStore(shortTerm, longTerm)

	router := model.NewRouter(
		map[string]model.ModelProvider{
			cfg.Model.Default:  model.NewOpenAIProvider(cfg.Model.Default),
			cfg.Model.Fallback: model.NewOpenAIProvider(cfg.Model.Fallback),
		},
		cfg.Model.Default,
		cfg.Model.Fallback,
	)

	policies := guardrails.NewPolicyEngine(cfg.Security.RequireApproval)
	engine := agent.NewEngine(executor, combined, router, policies, bus)
	controller := agent.NewController(engine)

	return cfg, runtimeDeps{
		Agent: controller,
		Close: func() error {
			return longTerm.Close()
		},
	}, nil
}

func runWithContext(ctx context.Context, runtime runtimeDeps, input string) (string, error) {
	return runtime.Agent.Run(ctx, input)
}
