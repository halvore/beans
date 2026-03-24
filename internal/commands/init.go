package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/hmans/beans/internal/gitutil"
	"github.com/hmans/beans/internal/output"
	"github.com/hmans/beans/pkg/beancore"
	"github.com/hmans/beans/pkg/config"
	"github.com/hmans/beans/pkg/localregistry"
)

var (
	initJSON  bool
	initLocal bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a beans project",
	Long: `Creates a .beans directory and .beans.yml config file in the current directory.

Use --local to store beans outside the project directory, in a local registry
at ~/.local/beans. The project directory will not be modified.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if initLocal {
			return initLocalProject()
		}

		var beansDir string

		if beansPath != "" {
			// Use explicit path for beans directory
			beansDir = beansPath
			// Create the directory using Core.Init to set up .gitignore
			core := beancore.New(beansDir, nil)
			if err := core.Init(); err != nil {
				if initJSON {
					return output.Error(output.ErrFileError, err.Error())
				}
				return fmt.Errorf("failed to create directory: %w", err)
			}
			// Skip creating .beans.yml when --beans-path is explicit:
			// the path is already known, and writing a config to the parent
			// directory could pollute unrelated locations (e.g. /tmp).
		} else {
			// Use current working directory
			dir, err := os.Getwd()
			if err != nil {
				if initJSON {
					return output.Error(output.ErrFileError, err.Error())
				}
				return err
			}

			if err := beancore.Init(dir); err != nil {
				if initJSON {
					return output.Error(output.ErrFileError, err.Error())
				}
				return fmt.Errorf("failed to initialize: %w", err)
			}

			projectDir := dir
			beansDir = filepath.Join(dir, ".beans")
			dirName := filepath.Base(dir)

			// Create default config file with directory name as prefix
			// Config is saved at project root (not inside .beans/)
			defaultCfg := config.DefaultWithPrefix(dirName + "-")
			defaultCfg.Project.Name = dirName
			defaultCfg.SetConfigDir(projectDir)

			// Auto-detect the remote's default branch if we're in a git repo
			if baseRef, ok := gitutil.DefaultRemoteBranch(projectDir, "origin"); ok {
				defaultCfg.Worktree.BaseRef = baseRef
			}
			if err := defaultCfg.Save(projectDir); err != nil {
				if initJSON {
					return output.Error(output.ErrFileError, err.Error())
				}
				return fmt.Errorf("failed to create config: %w", err)
			}
		}

		// Install default agent skills.
		installDefaultSkills(beansDir, false)

		// Install Claude Code command stubs for skill discoverability.
		if cwd, err := os.Getwd(); err == nil {
			commandsDir := filepath.Join(cwd, ".claude", "commands")
			installClaudeCodeCommands(commandsDir, filepath.Join(beansDir, "skills"), false)
		}

		if initJSON {
			return output.SuccessInit(beansDir)
		}

		fmt.Println("Initialized beans project")
		return nil
	},
}

// initLocalProject handles `beans init --local`. It registers the project in
// the local registry and stores beans + config outside the project directory.
func initLocalProject() error {
	dir, err := os.Getwd()
	if err != nil {
		if initJSON {
			return output.Error(output.ErrFileError, err.Error())
		}
		return err
	}

	dirName := filepath.Base(dir)

	// Load or create the local registry.
	reg, err := localregistry.Load()
	if err != nil {
		if initJSON {
			return output.Error(output.ErrFileError, err.Error())
		}
		return fmt.Errorf("failed to load local registry: %w", err)
	}

	// Register the project (idempotent — returns existing entry if already registered).
	entry, err := reg.Register(dir, dirName)
	if err != nil {
		if initJSON {
			return output.Error(output.ErrFileError, err.Error())
		}
		return fmt.Errorf("failed to register project: %w", err)
	}

	// Save the registry to disk.
	if err := reg.Save(); err != nil {
		if initJSON {
			return output.Error(output.ErrFileError, err.Error())
		}
		return fmt.Errorf("failed to save registry: %w", err)
	}

	// Initialize the .beans directory (creates .gitignore etc.) inside the local project dir.
	beansDir, err := reg.ProjectBeansDir(entry.Slug)
	if err != nil {
		if initJSON {
			return output.Error(output.ErrFileError, err.Error())
		}
		return fmt.Errorf("failed to resolve beans directory: %w", err)
	}

	core := beancore.New(beansDir, nil)
	if err := core.Init(); err != nil {
		if initJSON {
			return output.Error(output.ErrFileError, err.Error())
		}
		return fmt.Errorf("failed to initialize beans directory: %w", err)
	}

	// Install default agent skills.
	installDefaultSkills(beansDir, false)

	// Install Claude Code command stubs. For local projects, use $HOME/.claude/skills/
	// so we don't modify the project directory.
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to resolve home directory: %w", err)
	}
	localCommandsDir := filepath.Join(home, ".claude", "skills")
	installClaudeCodeCommands(localCommandsDir, filepath.Join(beansDir, "skills"), false)

	// Save config alongside the local beans dir.
	projectDir, err := reg.ProjectDir(entry.Slug)
	if err != nil {
		if initJSON {
			return output.Error(output.ErrFileError, err.Error())
		}
		return fmt.Errorf("failed to resolve project directory: %w", err)
	}

	defaultCfg := config.DefaultWithPrefix(dirName + "-")
	defaultCfg.Project.Name = dirName
	defaultCfg.SetConfigDir(projectDir)

	// Auto-detect the remote's default branch if we're in a git repo.
	if baseRef, ok := gitutil.DefaultRemoteBranch(dir, "origin"); ok {
		defaultCfg.Worktree.BaseRef = baseRef
	}

	if err := defaultCfg.Save(projectDir); err != nil {
		if initJSON {
			return output.Error(output.ErrFileError, err.Error())
		}
		return fmt.Errorf("failed to create config: %w", err)
	}

	if initJSON {
		return output.SuccessInit(beansDir)
	}

	fmt.Printf("Initialized local beans project at %s\n", projectDir)
	return nil
}

func RegisterInitCmd(root *cobra.Command) {
	initCmd.Flags().BoolVar(&initJSON, "json", false, "Output as JSON")
	initCmd.Flags().BoolVar(&initLocal, "local", false, "Store beans outside the project directory in a local registry")
	root.AddCommand(initCmd)
}
