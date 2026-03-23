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
		return nil
	},
}

func RegisterSkillsCmd(root *cobra.Command) {
	skillsInitCmd.Flags().BoolVar(&skillsInitForce, "force", false, "Overwrite existing skill files")
	skillsCmd.AddCommand(skillsInitCmd)
	root.AddCommand(skillsCmd)
}
