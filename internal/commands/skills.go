package commands

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hmans/beans/pkg/config"
	"github.com/spf13/cobra"
)

//go:embed default_skills/*.md
var defaultSkillsFS embed.FS

// installSkills writes the embedded default skill files directly into the
// given target directory. For in-repo projects this is typically
// <projectDir>/.claude/commands/; for local projects it is $HOME/.claude/skills/.
// Existing files are not overwritten unless force is true.
func installSkills(targetDir string, force bool) (int, error) {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return 0, fmt.Errorf("failed to create skills directory %s: %w", targetDir, err)
	}

	entries, err := defaultSkillsFS.ReadDir("default_skills")
	if err != nil {
		return 0, fmt.Errorf("failed to read embedded skills: %w", err)
	}

	installed := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		destPath := filepath.Join(targetDir, entry.Name())

		if !force {
			if _, err := os.Stat(destPath); err == nil {
				continue // skip existing
			}
		}

		data, err := defaultSkillsFS.ReadFile("default_skills/" + entry.Name())
		if err != nil {
			return installed, fmt.Errorf("failed to read embedded skill %s: %w", entry.Name(), err)
		}

		if err := os.WriteFile(destPath, data, 0644); err != nil {
			return installed, fmt.Errorf("failed to write skill %s: %w", entry.Name(), err)
		}
		installed++
	}

	return installed, nil
}

// claudeCommandsDir returns the directory where skills should be installed.
// For local projects (where ConfigDir differs from ProjectRoot), skills go
// to $HOME/.claude/skills/ to avoid modifying the project directory.
// For in-repo projects, they go to <projectDir>/.claude/commands/.
func claudeCommandsDir(c *config.Config, projectDir string) string {
	if c != nil && c.ConfigDir() != "" && c.ProjectRoot() != "" && c.ConfigDir() != c.ProjectRoot() {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, ".claude", "skills")
		}
	}
	return filepath.Join(projectDir, ".claude", "commands")
}

var skillsInitForce bool

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Manage agent skills",
}

var skillsInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Install default agent skills",
	Long: `Installs the default agent skill files into the Claude Code commands directory.
For in-repo projects this is .claude/commands/; for local projects it is $HOME/.claude/skills/.
Existing skill files are preserved unless --force is used.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// cfg is set by PersistentPreRunE in root.go
		projectDir := cfg.ProjectRoot()
		if projectDir == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return nil // best-effort
			}
			projectDir = cwd
		}

		targetDir := claudeCommandsDir(cfg, projectDir)
		installed, err := installSkills(targetDir, skillsInitForce)
		if err != nil {
			return err
		}

		if installed == 0 {
			fmt.Println("All default skills already installed (use --force to overwrite)")
		} else {
			fmt.Printf("Installed %d default skill(s) to %s\n", installed, targetDir)
		}

		return nil
	},
}

func RegisterSkillsCmd(root *cobra.Command) {
	skillsInitCmd.Flags().BoolVar(&skillsInitForce, "force", false, "Overwrite existing skill files")
	skillsCmd.AddCommand(skillsInitCmd)
	root.AddCommand(skillsCmd)
}
