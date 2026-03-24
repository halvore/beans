package commands

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

//go:embed default_skills/*.md
var defaultSkillsFS embed.FS

// installSkills writes the embedded default skill files directly into the
// given target directory. Existing files are not overwritten unless force is true.
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

// agentTool represents a supported AI coding agent tool.
type agentTool struct {
	Name    string // display name (e.g. "Claude")
	DirName string // home directory name (e.g. ".claude")
}

var supportedTools = []agentTool{
	{Name: "Claude", DirName: ".claude"},
	{Name: "Codex", DirName: ".codex"},
}

// detectTools returns the tools that appear to be installed by checking
// for their home config directory (~/.claude, ~/.codex, etc.).
func detectTools(home string) []agentTool {
	var detected []agentTool
	for _, tool := range supportedTools {
		dir := filepath.Join(home, tool.DirName)
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			detected = append(detected, tool)
		}
	}
	return detected
}

// skillsDir returns the beans skills directory for a given tool.
// e.g. $HOME/.claude/skills/beans
func skillsDir(home string, tool agentTool) string {
	return filepath.Join(home, tool.DirName, "skills", "beans")
}

// installSkillsForTools installs skills for the given tools and prints results.
func installSkillsForTools(home string, tools []agentTool, force bool) error {
	for _, tool := range tools {
		dir := skillsDir(home, tool)
		installed, err := installSkills(dir, force)
		if err != nil {
			return err
		}
		if installed == 0 {
			fmt.Printf("%s: all skills already installed (use --force to overwrite)\n", tool.Name)
		} else {
			fmt.Printf("%s: installed %d skill(s) to %s\n", tool.Name, installed, dir)
		}
	}
	return nil
}

var (
	skillsInitForce bool
	skillsInitClaude bool
	skillsInitCodex  bool
)

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Manage agent skills",
}

var skillsInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Install default agent skills",
	Long: `Installs the default agent skill files to your home directory.

By default, auto-detects installed tools (Claude, Codex) and installs for all
detected tools. Use --claude or --codex to install for specific tools only.

Skills are installed to:
  Claude: ~/.claude/skills/beans/
  Codex:  ~/.codex/skills/beans/

Existing skill files are preserved unless --force is used.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to resolve home directory: %w", err)
		}

		// Determine which tools to install for.
		var tools []agentTool
		explicitChoice := skillsInitClaude || skillsInitCodex
		if explicitChoice {
			if skillsInitClaude {
				tools = append(tools, agentTool{Name: "Claude", DirName: ".claude"})
			}
			if skillsInitCodex {
				tools = append(tools, agentTool{Name: "Codex", DirName: ".codex"})
			}
		} else {
			// Auto-detect installed tools.
			tools = detectTools(home)
			if len(tools) == 0 {
				fmt.Println("No supported tools detected (checked for ~/.claude and ~/.codex).")
				fmt.Println("Use --claude or --codex to install for a specific tool.")
				return nil
			}
			names := make([]string, len(tools))
			for i, t := range tools {
				names[i] = t.Name
			}
			fmt.Printf("Detected: %s\n", joinNames(names))
		}

		return installSkillsForTools(home, tools, skillsInitForce)
	},
}

// joinNames joins names with commas and "and" for the last element.
func joinNames(names []string) string {
	switch len(names) {
	case 0:
		return ""
	case 1:
		return names[0]
	case 2:
		return names[0] + " and " + names[1]
	default:
		result := ""
		for i, n := range names {
			if i == len(names)-1 {
				result += "and " + n
			} else {
				result += n + ", "
			}
		}
		return result
	}
}

func RegisterSkillsCmd(root *cobra.Command) {
	skillsInitCmd.Flags().BoolVar(&skillsInitForce, "force", false, "Overwrite existing skill files")
	skillsInitCmd.Flags().BoolVar(&skillsInitClaude, "claude", false, "Install skills for Claude")
	skillsInitCmd.Flags().BoolVar(&skillsInitCodex, "codex", false, "Install skills for Codex")
	skillsCmd.AddCommand(skillsInitCmd)
	root.AddCommand(skillsCmd)
}
