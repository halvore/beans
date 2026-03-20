package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/hmans/beans/internal/agent"
	"github.com/hmans/beans/internal/tui"
	"github.com/hmans/beans/internal/worktree"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Open the interactive TUI",
	Long:  `Opens an interactive terminal user interface for browsing and managing beans.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create agent manager with a context provider that returns bean info
		agentMgr := agent.NewManager(core.Root(), func(beanID string) string {
			b, err := core.Get(beanID)
			if err != nil {
				return ""
			}
			var ctx string
			ctx = fmt.Sprintf("You are working on bean %s: %q\n", b.ID, b.Title)
			ctx += fmt.Sprintf("Type: %s | Status: %s", b.Type, b.Status)
			if b.Priority != "" {
				ctx += fmt.Sprintf(" | Priority: %s", b.Priority)
			}
			ctx += "\n"
			if b.Body != "" {
				ctx += fmt.Sprintf("\nDescription:\n%s", b.Body)
			}
			return ctx
		}, agent.DefaultMode(cfg.GetDefaultMode()))

		// Optionally create worktree manager for multi-agent mode
		var wtMgr *worktree.Manager
		if cfg.IsWorktreeMode() {
			projectRoot := cfg.ProjectRoot()
			if projectRoot == "" {
				projectRoot = filepath.Dir(core.Root())
			}
			projectName := cfg.GetProjectName()
			if projectName == "" {
				projectName = filepath.Base(projectRoot)
			}
			worktreeRoot, err := cfg.ResolveWorktreePath(projectName)
			if err != nil {
				return fmt.Errorf("failed to resolve worktree path: %w", err)
			}
			if err := os.MkdirAll(worktreeRoot, 0755); err != nil {
				return fmt.Errorf("failed to create worktree directory: %w", err)
			}
			wtMgr = worktree.NewManager(projectRoot, worktreeRoot, cfg.GetWorktreeBaseRef(), cfg.GetWorktreeSetup())
		}

		return tui.Run(core, cfg, agentMgr, wtMgr)
	},
}

func RegisterTuiCmd(root *cobra.Command) {
	root.AddCommand(tuiCmd)
}
