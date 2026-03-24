package commands

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed default_skills/*.md
var defaultSkillsFS embed.FS

// installDefaultSkills writes the embedded default skill files into the
// skills/ subdirectory of the given beansPath. Existing files are not
// overwritten unless force is true.
func installDefaultSkills(beansPath string, force bool) (int, error) {
	skillsDir := filepath.Join(beansPath, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return 0, fmt.Errorf("failed to create skills directory: %w", err)
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

		destPath := filepath.Join(skillsDir, entry.Name())

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

// installClaudeCodeCommands creates stub command files in the project's
// .claude/commands/ directory so that Claude Code's slash command system
// discovers beans skills. Each stub reads the full skill file from the
// actual skills directory (which may be outside the project for local storage).
func installClaudeCodeCommands(projectDir, skillsDir string, force bool) (int, error) {
	commandsDir := filepath.Join(projectDir, ".claude", "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		return 0, fmt.Errorf("failed to create .claude/commands directory: %w", err)
	}

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return 0, nil // no skills dir, nothing to do
	}

	installed := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		destPath := filepath.Join(commandsDir, entry.Name())

		if !force {
			if _, err := os.Stat(destPath); err == nil {
				continue // skip existing
			}
		}

		skillPath := filepath.Join(skillsDir, entry.Name())
		stub := fmt.Sprintf("Read and follow the full skill instructions in `%s`.\n", skillPath)

		if err := os.WriteFile(destPath, []byte(stub), 0644); err != nil {
			return installed, fmt.Errorf("failed to write command stub %s: %w", entry.Name(), err)
		}
		installed++
	}

	return installed, nil
}

var skillsInitForce bool

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Manage agent skills",
}

var skillsInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Install default agent skills",
	Long: `Installs the default agent skill files into the .beans/skills/ directory.
Existing skill files are preserved unless --force is used.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// cfg is set by PersistentPreRunE in root.go
		bp := cfg.ResolveBeansPath()

		installed, err := installDefaultSkills(bp, skillsInitForce)
		if err != nil {
			return err
		}

		if installed == 0 {
			fmt.Println("All default skills already installed (use --force to overwrite)")
		} else {
			fmt.Printf("Installed %d default skill(s) to %s\n", installed, filepath.Join(bp, "skills"))
		}

		// Install Claude Code command stubs so skills are discoverable via slash commands.
		projectDir := cfg.ProjectRoot()
		if projectDir == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return nil // best-effort
			}
			projectDir = cwd
		}
		skillsDir := filepath.Join(bp, "skills")
		ccInstalled, err := installClaudeCodeCommands(projectDir, skillsDir, skillsInitForce)
		if err != nil {
			return err
		}
		if ccInstalled > 0 {
			fmt.Printf("Installed %d Claude Code command(s) to %s\n", ccInstalled, filepath.Join(projectDir, ".claude", "commands"))
		}

		return nil
	},
}

func RegisterSkillsCmd(root *cobra.Command) {
	skillsInitCmd.Flags().BoolVar(&skillsInitForce, "force", false, "Overwrite existing skill files")
	skillsCmd.AddCommand(skillsInitCmd)
	root.AddCommand(skillsCmd)
}
