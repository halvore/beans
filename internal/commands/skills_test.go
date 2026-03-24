package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
		projectDir := t.TempDir()
		skillsDir := filepath.Join(t.TempDir(), "skills")
		os.MkdirAll(skillsDir, 0755)

		os.WriteFile(filepath.Join(skillsDir, "bplan.md"), []byte("# /bplan — Critical Bean Planning"), 0644)
		os.WriteFile(filepath.Join(skillsDir, "breview.md"), []byte("# /breview — Pre-PR Code Review"), 0644)

		installed, err := installClaudeCodeCommands(projectDir, skillsDir, false)
		if err != nil {
			t.Fatalf("installClaudeCodeCommands() error = %v", err)
		}
		if installed != 2 {
			t.Errorf("installed = %d, want 2", installed)
		}

		// Verify stubs exist and reference the correct skill path.
		for _, name := range []string{"bplan.md", "breview.md"} {
			stubPath := filepath.Join(projectDir, ".claude", "commands", name)
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
		projectDir := t.TempDir()
		commandsDir := filepath.Join(projectDir, ".claude", "commands")
		os.MkdirAll(commandsDir, 0755)

		custom := []byte("Custom command content")
		os.WriteFile(filepath.Join(commandsDir, "bplan.md"), custom, 0644)

		skillsDir := filepath.Join(t.TempDir(), "skills")
		os.MkdirAll(skillsDir, 0755)
		os.WriteFile(filepath.Join(skillsDir, "bplan.md"), []byte("# skill"), 0644)

		installed, err := installClaudeCodeCommands(projectDir, skillsDir, false)
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
		projectDir := t.TempDir()
		commandsDir := filepath.Join(projectDir, ".claude", "commands")
		os.MkdirAll(commandsDir, 0755)
		os.WriteFile(filepath.Join(commandsDir, "bplan.md"), []byte("old"), 0644)

		skillsDir := filepath.Join(t.TempDir(), "skills")
		os.MkdirAll(skillsDir, 0755)
		os.WriteFile(filepath.Join(skillsDir, "bplan.md"), []byte("# skill"), 0644)

		installed, err := installClaudeCodeCommands(projectDir, skillsDir, true)
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
		projectDir := t.TempDir()
		installed, err := installClaudeCodeCommands(projectDir, "/nonexistent/skills", false)
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if installed != 0 {
			t.Errorf("installed = %d, want 0", installed)
		}
	})
}
