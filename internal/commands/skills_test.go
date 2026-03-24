package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hmans/beans/pkg/config"
)

func TestInstallDefaultSkills(t *testing.T) {
	t.Run("installs all default skills", func(t *testing.T) {
		beansDir := t.TempDir()

		installed, err := installDefaultSkills(beansDir, false)
		if err != nil {
			t.Fatalf("installDefaultSkills() error = %v", err)
		}
		if installed != 4 {
			t.Errorf("installed = %d, want 4", installed)
		}

		// Verify files exist
		for _, name := range []string{"bplan.md", "breview.md", "bship.md", "binvestigate.md"} {
			path := filepath.Join(beansDir, "skills", name)
			if _, err := os.Stat(path); err != nil {
				t.Errorf("expected skill file %s to exist", name)
			}
		}
	})

	t.Run("does not overwrite existing files", func(t *testing.T) {
		beansDir := t.TempDir()
		skillsDir := filepath.Join(beansDir, "skills")
		os.MkdirAll(skillsDir, 0755)

		// Write a custom bplan.md
		customContent := []byte("# Custom plan skill")
		os.WriteFile(filepath.Join(skillsDir, "bplan.md"), customContent, 0644)

		installed, err := installDefaultSkills(beansDir, false)
		if err != nil {
			t.Fatalf("installDefaultSkills() error = %v", err)
		}
		// Should install 3 (skipping bplan.md)
		if installed != 3 {
			t.Errorf("installed = %d, want 3", installed)
		}

		// Verify custom file was preserved
		data, _ := os.ReadFile(filepath.Join(skillsDir, "bplan.md"))
		if string(data) != string(customContent) {
			t.Error("custom bplan.md was overwritten")
		}
	})

	t.Run("force overwrites existing files", func(t *testing.T) {
		beansDir := t.TempDir()
		skillsDir := filepath.Join(beansDir, "skills")
		os.MkdirAll(skillsDir, 0755)

		// Write a custom bplan.md
		os.WriteFile(filepath.Join(skillsDir, "bplan.md"), []byte("# Custom"), 0644)

		installed, err := installDefaultSkills(beansDir, true)
		if err != nil {
			t.Fatalf("installDefaultSkills() error = %v", err)
		}
		if installed != 4 {
			t.Errorf("installed = %d, want 4", installed)
		}

		// Verify custom file was overwritten
		data, _ := os.ReadFile(filepath.Join(skillsDir, "bplan.md"))
		if string(data) == "# Custom" {
			t.Error("custom bplan.md was NOT overwritten with force=true")
		}
	})
}

func TestInstallClaudeCodeCommands(t *testing.T) {
	t.Run("creates command stubs pointing to skill files", func(t *testing.T) {
		targetDir := filepath.Join(t.TempDir(), ".claude", "commands")
		skillsDir := filepath.Join(t.TempDir(), "skills")
		os.MkdirAll(skillsDir, 0755)

		os.WriteFile(filepath.Join(skillsDir, "bplan.md"), []byte("# /bplan — Critical Bean Planning"), 0644)
		os.WriteFile(filepath.Join(skillsDir, "breview.md"), []byte("# /breview — Pre-PR Code Review"), 0644)

		installed, err := installClaudeCodeCommands(targetDir, skillsDir, false)
		if err != nil {
			t.Fatalf("installClaudeCodeCommands() error = %v", err)
		}
		if installed != 2 {
			t.Errorf("installed = %d, want 2", installed)
		}

		// Verify stubs exist and reference the correct skill path.
		for _, name := range []string{"bplan.md", "breview.md"} {
			stubPath := filepath.Join(targetDir, name)
			data, err := os.ReadFile(stubPath)
			if err != nil {
				t.Errorf("expected stub %s to exist: %v", name, err)
				continue
			}
			expectedPath := filepath.Join(skillsDir, name)
			if !strings.Contains(string(data), expectedPath) {
				t.Errorf("stub %s should reference %s, got: %s", name, expectedPath, string(data))
			}
		}
	})

	t.Run("does not overwrite existing stubs", func(t *testing.T) {
		commandsDir := filepath.Join(t.TempDir(), ".claude", "commands")
		os.MkdirAll(commandsDir, 0755)

		custom := []byte("Custom command content")
		os.WriteFile(filepath.Join(commandsDir, "bplan.md"), custom, 0644)

		skillsDir := filepath.Join(t.TempDir(), "skills")
		os.MkdirAll(skillsDir, 0755)
		os.WriteFile(filepath.Join(skillsDir, "bplan.md"), []byte("# skill"), 0644)

		installed, err := installClaudeCodeCommands(commandsDir, skillsDir, false)
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if installed != 0 {
			t.Errorf("installed = %d, want 0 (should skip existing)", installed)
		}

		data, _ := os.ReadFile(filepath.Join(commandsDir, "bplan.md"))
		if string(data) != string(custom) {
			t.Error("existing command file was overwritten")
		}
	})

	t.Run("force overwrites existing stubs", func(t *testing.T) {
		commandsDir := filepath.Join(t.TempDir(), ".claude", "commands")
		os.MkdirAll(commandsDir, 0755)
		os.WriteFile(filepath.Join(commandsDir, "bplan.md"), []byte("old"), 0644)

		skillsDir := filepath.Join(t.TempDir(), "skills")
		os.MkdirAll(skillsDir, 0755)
		os.WriteFile(filepath.Join(skillsDir, "bplan.md"), []byte("# skill"), 0644)

		installed, err := installClaudeCodeCommands(commandsDir, skillsDir, true)
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if installed != 1 {
			t.Errorf("installed = %d, want 1", installed)
		}

		data, _ := os.ReadFile(filepath.Join(commandsDir, "bplan.md"))
		if string(data) == "old" {
			t.Error("stub was NOT overwritten with force=true")
		}
	})

	t.Run("returns zero for missing skills directory", func(t *testing.T) {
		targetDir := filepath.Join(t.TempDir(), "commands")
		installed, err := installClaudeCodeCommands(targetDir, "/nonexistent/skills", false)
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if installed != 0 {
			t.Errorf("installed = %d, want 0", installed)
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
