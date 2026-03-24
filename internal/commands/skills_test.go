package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hmans/beans/pkg/config"
)

func TestInstallSkills(t *testing.T) {
	t.Run("installs all default skills", func(t *testing.T) {
		targetDir := filepath.Join(t.TempDir(), ".claude", "commands")

		installed, err := installSkills(targetDir, false)
		if err != nil {
			t.Fatalf("installSkills() error = %v", err)
		}
		if installed != 4 {
			t.Errorf("installed = %d, want 4", installed)
		}

		// Verify files exist
		for _, name := range []string{"bplan.md", "breview.md", "bship.md", "binvestigate.md"} {
			path := filepath.Join(targetDir, name)
			if _, err := os.Stat(path); err != nil {
				t.Errorf("expected skill file %s to exist", name)
			}

			// Verify full content was written (not a stub)
			data, _ := os.ReadFile(path)
			if len(data) < 100 {
				t.Errorf("skill %s seems too short (%d bytes), expected full content", name, len(data))
			}
		}
	})

	t.Run("does not overwrite existing files", func(t *testing.T) {
		targetDir := filepath.Join(t.TempDir(), ".claude", "commands")
		os.MkdirAll(targetDir, 0755)

		// Write a custom bplan.md
		customContent := []byte("# Custom plan skill")
		os.WriteFile(filepath.Join(targetDir, "bplan.md"), customContent, 0644)

		installed, err := installSkills(targetDir, false)
		if err != nil {
			t.Fatalf("installSkills() error = %v", err)
		}
		// Should install 3 (skipping bplan.md)
		if installed != 3 {
			t.Errorf("installed = %d, want 3", installed)
		}

		// Verify custom file was preserved
		data, _ := os.ReadFile(filepath.Join(targetDir, "bplan.md"))
		if string(data) != string(customContent) {
			t.Error("custom bplan.md was overwritten")
		}
	})

	t.Run("force overwrites existing files", func(t *testing.T) {
		targetDir := filepath.Join(t.TempDir(), ".claude", "commands")
		os.MkdirAll(targetDir, 0755)

		// Write a custom bplan.md
		os.WriteFile(filepath.Join(targetDir, "bplan.md"), []byte("# Custom"), 0644)

		installed, err := installSkills(targetDir, true)
		if err != nil {
			t.Fatalf("installSkills() error = %v", err)
		}
		if installed != 4 {
			t.Errorf("installed = %d, want 4", installed)
		}

		// Verify custom file was overwritten
		data, _ := os.ReadFile(filepath.Join(targetDir, "bplan.md"))
		if string(data) == "# Custom" {
			t.Error("custom bplan.md was NOT overwritten with force=true")
		}
	})
}

func TestClaudeCommandsDir(t *testing.T) {
	t.Run("returns project .claude/commands for in-repo projects", func(t *testing.T) {
		c := &config.Config{}
		projectDir := "/some/project"
		c.SetConfigDir(projectDir)

		got := claudeCommandsDir(c, projectDir)
		want := filepath.Join(projectDir, ".claude", "commands")
		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	})

	t.Run("returns home .claude/skills for local projects", func(t *testing.T) {
		c := &config.Config{}
		c.SetConfigDir("/home/user/.local/beans/projects/myproject")
		c.SetProjectRoot("/some/project")

		got := claudeCommandsDir(c, "/some/project")
		home, _ := os.UserHomeDir()
		want := filepath.Join(home, ".claude", "skills")
		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	})

	t.Run("returns project .claude/commands for nil config", func(t *testing.T) {
		projectDir := "/some/project"
		got := claudeCommandsDir(nil, projectDir)
		want := filepath.Join(projectDir, ".claude", "commands")
		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	})
}
