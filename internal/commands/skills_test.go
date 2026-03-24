package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallSkills(t *testing.T) {
	t.Run("installs all default skills", func(t *testing.T) {
		targetDir := filepath.Join(t.TempDir(), ".claude", "skills", "beans")

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
		targetDir := filepath.Join(t.TempDir(), ".claude", "skills", "beans")
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
		targetDir := filepath.Join(t.TempDir(), ".claude", "skills", "beans")
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

func TestDetectTools(t *testing.T) {
	t.Run("detects claude directory", func(t *testing.T) {
		home := t.TempDir()
		os.MkdirAll(filepath.Join(home, ".claude"), 0755)

		tools := detectTools(home)
		if len(tools) != 1 {
			t.Fatalf("detected %d tools, want 1", len(tools))
		}
		if tools[0].Name != "Claude" {
			t.Errorf("detected tool = %s, want Claude", tools[0].Name)
		}
	})

	t.Run("detects both claude and codex", func(t *testing.T) {
		home := t.TempDir()
		os.MkdirAll(filepath.Join(home, ".claude"), 0755)
		os.MkdirAll(filepath.Join(home, ".codex"), 0755)

		tools := detectTools(home)
		if len(tools) != 2 {
			t.Fatalf("detected %d tools, want 2", len(tools))
		}
	})

	t.Run("returns empty when nothing detected", func(t *testing.T) {
		home := t.TempDir()

		tools := detectTools(home)
		if len(tools) != 0 {
			t.Errorf("detected %d tools, want 0", len(tools))
		}
	})
}

func TestSkillsDir(t *testing.T) {
	home := "/home/user"
	tool := agentTool{Name: "Claude", DirName: ".claude"}

	got := skillsDir(home, tool)
	want := filepath.Join(home, ".claude", "skills", "beans")
	if got != want {
		t.Errorf("skillsDir() = %s, want %s", got, want)
	}
}

func TestInstallSkillsForTools(t *testing.T) {
	t.Run("installs to both tools", func(t *testing.T) {
		home := t.TempDir()
		os.MkdirAll(filepath.Join(home, ".claude"), 0755)
		os.MkdirAll(filepath.Join(home, ".codex"), 0755)

		tools := []agentTool{
			{Name: "Claude", DirName: ".claude"},
			{Name: "Codex", DirName: ".codex"},
		}

		err := installSkillsForTools(home, tools, false)
		if err != nil {
			t.Fatalf("installSkillsForTools() error = %v", err)
		}

		// Verify skills exist in both directories
		for _, tool := range tools {
			dir := skillsDir(home, tool)
			for _, name := range []string{"bplan.md", "breview.md", "bship.md", "binvestigate.md"} {
				path := filepath.Join(dir, name)
				if _, err := os.Stat(path); err != nil {
					t.Errorf("expected skill file %s to exist for %s", name, tool.Name)
				}
			}
		}
	})
}

func TestJoinNames(t *testing.T) {
	tests := []struct {
		names []string
		want  string
	}{
		{nil, ""},
		{[]string{"Claude"}, "Claude"},
		{[]string{"Claude", "Codex"}, "Claude and Codex"},
		{[]string{"A", "B", "C"}, "A, B, and C"},
	}

	for _, tt := range tests {
		got := joinNames(tt.names)
		if got != tt.want {
			t.Errorf("joinNames(%v) = %q, want %q", tt.names, got, tt.want)
		}
	}
}
