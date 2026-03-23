package commands

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/hmans/beans/pkg/config"
	"github.com/spf13/cobra"
)

//go:embed prompt.tmpl
var agentPromptTemplate string

const notInitializedPrompt = `<EXTREMELY_IMPORTANT>
# Beans Is Not Initialized

This project does not have beans set up yet. Before you can use beans to track work, you MUST ask the user which storage mode to use:

1. **In-repo storage** (` + "`beans init`" + `): Stores beans as markdown files in a ` + "`.beans/`" + ` directory inside the repo. This is the default and recommended for most projects. Bean files are committed to the repo alongside the code.
2. **Local storage** (` + "`beans init --local`" + `): Stores beans outside the repo in a local directory. Use this when you don't want bean files in the repo.

**You MUST ask the user which option they prefer before running either command. Do NOT choose for them.**

Once the user has chosen, run the appropriate command, and then re-run ` + "`beans prime`" + ` to get the full usage guide.
</EXTREMELY_IMPORTANT>
`

// skillInfo holds the name and first-line description of a skill file.
type skillInfo struct {
	Name        string
	Description string
}

// promptData holds all data needed to render the prompt template.
type promptData struct {
	GraphQLSchema string
	Types         []config.TypeConfig
	Statuses      []config.StatusConfig
	Priorities    []config.PriorityConfig
	Skills        []skillInfo
	SkillsDir     string // Absolute path to the skills directory
}

// discoverSkills reads .md files from the skills directory under beansPath
// and extracts the skill name (from filename) and description (from the
// first heading or first non-empty line).
func discoverSkills(beansPath string) []skillInfo {
	skillsDir := filepath.Join(beansPath, "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil
	}

	var skills []skillInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".md")
		desc := extractSkillDescription(filepath.Join(skillsDir, entry.Name()))

		skills = append(skills, skillInfo{
			Name:        name,
			Description: desc,
		})
	}
	return skills
}

// extractSkillDescription reads the first heading from a skill file.
// It looks for a line starting with "# " and extracts the text after
// the " — " separator. Falls back to the full heading text, then to
// the filename.
func extractSkillDescription(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			heading := strings.TrimPrefix(line, "# ")
			// Look for " — " separator (e.g. "# /plan — Critical Bean Planning")
			if idx := strings.Index(heading, " — "); idx >= 0 {
				return strings.TrimSpace(heading[idx+len(" — "):])
			}
			return heading
		}
	}
	return ""
}

var primeCmd = &cobra.Command{
	Use:   "prime",
	Short: "Output instructions for AI coding agents",
	Long:  `Outputs a prompt that primes AI coding agents on how to use the beans CLI to manage project issues.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config for prime (skipped by PersistentPreRunE).
		// Check in-repo config first, then fall back to local registry.
		var primeCfg *config.Config
		if configPath != "" {
			var err error
			primeCfg, err = config.Load(configPath)
			if err != nil {
				return nil // Silently exit on error
			}
		} else {
			cwd, err := os.Getwd()
			if err != nil {
				return nil // Silently exit on error
			}

			configFile, err := config.FindConfig(cwd)
			if err != nil {
				return nil
			}

			if configFile != "" {
				primeCfg, err = config.Load(configFile)
				if err != nil {
					return nil
				}
			} else {
				// No in-repo config — check local registry
				primeCfg, err = loadFromLocalRegistry(cwd)
				if err != nil {
					return nil
				}
				// If no local registry entry was found, the config will have
				// default beans path. Check if a .beans dir actually exists.
				beansDir := primeCfg.ResolveBeansPath()
				if info, statErr := os.Stat(beansDir); statErr != nil || !info.IsDir() {
					fmt.Fprint(os.Stdout, notInitializedPrompt)
					return nil
				}
			}
		}

		tmpl, err := template.New("prompt").Parse(agentPromptTemplate)
		if err != nil {
			return err
		}

		beansPath := primeCfg.ResolveBeansPath()

		skillsDir := filepath.Join(beansPath, "skills")

		data := promptData{
			GraphQLSchema: GetGraphQLSchema(),
			Types:         config.DefaultTypes,
			Statuses:      config.DefaultStatuses,
			Priorities:    config.DefaultPriorities,
			Skills:        discoverSkills(beansPath),
			SkillsDir:     skillsDir,
		}

		return tmpl.Execute(os.Stdout, data)
	},
}

func RegisterPrimeCmd(root *cobra.Command) {
	root.AddCommand(primeCmd)
}
