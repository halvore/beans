package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hmans/beans/pkg/localregistry"
	"github.com/spf13/cobra"
)

func RegisterProjectsCmd(root *cobra.Command) {
	var listJSON bool
	var removeJSON bool
	var removeData bool

	projectsCmd := &cobra.Command{
		Use:   "projects",
		Short: "Manage locally registered projects",
		Long:  `Commands for managing projects registered in the local beans registry at ~/.local/beans.`,
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Short:   "List all locally registered projects",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := localregistry.Load()
			if err != nil {
				return fmt.Errorf("loading registry: %w", err)
			}

			if listJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(reg.Projects)
			}

			if len(reg.Projects) == 0 {
				fmt.Println("No projects registered.")
				fmt.Println("Use 'beans init --local' to register a project.")
				return nil
			}

			for _, p := range reg.Projects {
				fmt.Printf("%s\t%s\t%s\n", p.Slug, p.Path, p.RegisteredAt.Format("2006-01-02"))
			}
			return nil
		},
	}

	removeCmd := &cobra.Command{
		Use:   "remove <path>",
		Short: "Unregister a project from the local registry",
		Long: `Removes a project from the local registry. By default, the project's local
beans data is preserved. Use --delete-data to also remove the stored beans.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectPath := args[0]

			reg, err := localregistry.Load()
			if err != nil {
				return fmt.Errorf("loading registry: %w", err)
			}

			// Look up the entry before removing (we need the slug for --delete-data).
			entry := reg.Lookup(projectPath)
			if entry == nil {
				if removeJSON {
					return projectsJSONError("NOT_FOUND", fmt.Sprintf("project not found in registry: %s", projectPath))
				}
				return fmt.Errorf("project not found in registry: %s", projectPath)
			}

			slug := entry.Slug

			if !reg.Unregister(projectPath) {
				if removeJSON {
					return projectsJSONError("NOT_FOUND", fmt.Sprintf("project not found in registry: %s", projectPath))
				}
				return fmt.Errorf("project not found in registry: %s", projectPath)
			}

			if err := reg.Save(); err != nil {
				return fmt.Errorf("saving registry: %w", err)
			}

			if removeData {
				projectDir, err := reg.ProjectDir(slug)
				if err != nil {
					return fmt.Errorf("resolving project directory: %w", err)
				}
				if err := os.RemoveAll(projectDir); err != nil {
					return fmt.Errorf("deleting project data: %w", err)
				}
			}

			if removeJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]any{
					"success": true,
					"message": fmt.Sprintf("Unregistered project: %s", projectPath),
				})
			}

			fmt.Printf("Unregistered project: %s\n", projectPath)
			if !removeData {
				fmt.Println("Local beans data was preserved. Use --delete-data to also remove it.")
			}
			return nil
		},
	}

	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output as JSON")
	removeCmd.Flags().BoolVar(&removeJSON, "json", false, "Output as JSON")
	removeCmd.Flags().BoolVar(&removeData, "delete-data", false, "Also delete the project's local beans data")

	projectsCmd.AddCommand(listCmd)
	projectsCmd.AddCommand(removeCmd)
	root.AddCommand(projectsCmd)
}

func projectsJSONError(code, message string) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(map[string]any{
		"success": false,
		"error":   message,
		"code":    code,
	})
	return fmt.Errorf("%s", message)
}
