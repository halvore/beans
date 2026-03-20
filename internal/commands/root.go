package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hmans/beans/pkg/beancore"
	"github.com/hmans/beans/pkg/config"
	"github.com/hmans/beans/pkg/localregistry"
	"github.com/spf13/cobra"
)

var core *beancore.Core
var cfg *config.Config
var beansPath string
var configPath string

// NewRootCmd creates the root cobra command with shared persistent flags
// and core initialization logic.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "beans",
		Short: "A file-based issue tracker for AI-first workflows",
		Long: `Beans is a lightweight issue tracker that stores issues as markdown files.
Track your work alongside your code and supercharge your coding agent with
a full view of your project.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip core initialization for init, prime, and version commands
			if cmd.Name() == "init" || cmd.Name() == "prime" || cmd.Name() == "version" {
				return nil
			}

			var err error

			// Load configuration
			if configPath != "" {
				cfg, err = config.Load(configPath)
				if err != nil {
					return fmt.Errorf("loading config from %s: %w", configPath, err)
				}
			} else {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("getting current directory: %w", err)
				}

				// Try to find .beans.yml in the project tree
				configFile, err := config.FindConfig(cwd)
				if err != nil {
					return fmt.Errorf("finding config: %w", err)
				}

				if configFile != "" {
					cfg, err = config.Load(configFile)
					if err != nil {
						return fmt.Errorf("loading config: %w", err)
					}
				} else {
					// No config in project tree — check local registry
					cfg, err = loadFromLocalRegistry(cwd)
					if err != nil {
						return fmt.Errorf("loading config: %w", err)
					}
				}
			}

			root, err := resolveBeansPath(beansPath, cfg)
			if err != nil {
				return err
			}

			core = beancore.New(root, cfg)
			if err := core.Load(); err != nil {
				return fmt.Errorf("loading beans: %w", err)
			}

			return nil
		},
	}

	rootCmd.PersistentFlags().StringVar(&beansPath, "beans-path", "", "Path to data directory (overrides config and BEANS_PATH env var)")
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Path to config file (default: searches upward for .beans.yml)")

	return rootCmd
}

// resolveBeansPath determines the beans data directory path.
// Precedence: --beans-path flag > BEANS_PATH env var > config default.
//
// In worktrees, the CLI uses the worktree's local .beans/ directory.
// beans-serve watches worktree .beans/ dirs and merges changes into
// runtime state, so the UI stays up-to-date without writing to main.
func resolveBeansPath(flagPath string, c *config.Config) (string, error) {
	explicitOverride := flagPath != "" || os.Getenv("BEANS_PATH") != ""

	var root string
	if flagPath != "" {
		root = flagPath
	} else if envPath := os.Getenv("BEANS_PATH"); envPath != "" {
		root = envPath
	} else {
		root = c.ResolveBeansPath()
	}

	if info, statErr := os.Stat(root); statErr != nil || !info.IsDir() {
		if explicitOverride {
			return "", fmt.Errorf("beans path does not exist or is not a directory: %s", root)
		}
		return "", fmt.Errorf("no .beans directory found at %s (run 'beans init' to create one)", root)
	}

	return root, nil
}

// loadFromLocalRegistry checks the local registry for a project matching the
// given directory. If found, loads config from the local project directory.
// Returns a default config anchored at dir if not found in the registry.
func loadFromLocalRegistry(dir string) (*config.Config, error) {
	reg, err := localregistry.Load()
	if err != nil {
		// Registry doesn't exist or can't be read — fall back to default
		cfg := config.Default()
		cfg.SetConfigDir(dir)
		return cfg, nil
	}

	entry := reg.Lookup(dir)
	if entry == nil {
		cfg := config.Default()
		cfg.SetConfigDir(dir)
		return cfg, nil
	}

	projectDir, err := reg.ProjectDir(entry.Slug)
	if err != nil {
		return nil, fmt.Errorf("resolving local project directory: %w", err)
	}

	configFile := filepath.Join(projectDir, config.ConfigFileName)
	cfg, err := config.Load(configFile)
	if err != nil {
		return nil, err
	}
	// The actual project lives at entry.Path, not in the local registry dir.
	cfg.SetProjectRoot(entry.Path)
	return cfg, nil
}

// Execute runs the given root command and exits on error.
func Execute(rootCmd *cobra.Command) {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
