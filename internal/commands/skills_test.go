package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallSkillsFlat(t *testing.T) {
	t.Run("installs all default skills", func(t *testing.T) {
		targetDir := filepath.Join(t.TempDir(), ".codex", "skills", "beans")

		installed, err := installSkills(targetDir, skillFormatFlat, false)
		if err != nil {
			t.Fatalf("installSkills() error = %v", err)
		}
		if installed != 5 {
			t.Errorf("installed = %d, want 5", installed)
		}

		// Verify files exist
		for _, name := range []string{"bplan.md", "brefine.md", "breview.md", "bship.md", "binvestigate.md"} {
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
		targetDir := filepath.Join(t.TempDir(), ".codex", "skills", "beans")
		os.MkdirAll(targetDir, 0755)

		// Write a custom bplan.md
		customContent := []byte("# Custom plan skill")
		os.WriteFile(filepath.Join(targetDir, "bplan.md"), customContent, 0644)

		installed, err := installSkills(targetDir, skillFormatFlat, false)
		if err != nil {
			t.Fatalf("installSkills() error = %v", err)
		}
		// Should install 4 (skipping bplan.md)
		if installed != 4 {
			t.Errorf("installed = %d, want 4", installed)
		}

		// Verify custom file was preserved
		data, _ := os.ReadFile(filepath.Join(targetDir, "bplan.md"))
		if string(data) != string(customContent) {
			t.Error("custom bplan.md was overwritten")
		}
	})

	t.Run("force overwrites existing files", func(t *testing.T) {
		targetDir := filepath.Join(t.TempDir(), ".codex", "skills", "beans")
		os.MkdirAll(targetDir, 0755)

		// Write a custom bplan.md
		os.WriteFile(filepath.Join(targetDir, "bplan.md"), []byte("# Custom"), 0644)

		installed, err := installSkills(targetDir, skillFormatFlat, true)
		if err != nil {
			t.Fatalf("installSkills() error = %v", err)
		}
		if installed != 5 {
			t.Errorf("installed = %d, want 5", installed)
		}

		// Verify custom file was overwritten
		data, _ := os.ReadFile(filepath.Join(targetDir, "bplan.md"))
		if string(data) == "# Custom" {
			t.Error("custom bplan.md was NOT overwritten with force=true")
		}
	})
}

func TestInstallSkillsNative(t *testing.T) {
	t.Run("installs skills as SKILL.md in subdirectories", func(t *testing.T) {
		targetDir := filepath.Join(t.TempDir(), ".claude", "skills")

		installed, err := installSkills(targetDir, skillFormatNative, false)
		if err != nil {
			t.Fatalf("installSkills() error = %v", err)
		}
		if installed != 5 {
			t.Errorf("installed = %d, want 5", installed)
		}

		// Verify each skill is in its own subdirectory with SKILL.md
		for _, name := range []string{"bplan", "brefine", "breview", "bship", "binvestigate"} {
			skillDir := filepath.Join(targetDir, "beans-"+name)
			skillFile := filepath.Join(skillDir, "SKILL.md")

			if _, err := os.Stat(skillFile); err != nil {
				t.Errorf("expected SKILL.md at %s to exist", skillFile)
				continue
			}

			data, _ := os.ReadFile(skillFile)
			content := string(data)

			// Verify YAML frontmatter is present
			if !strings.HasPrefix(content, "---\n") {
				t.Errorf("skill %s missing YAML frontmatter", name)
			}
			if !strings.Contains(content, "name: "+name) {
				t.Errorf("skill %s missing name in frontmatter", name)
			}
			if !strings.Contains(content, "description: ") {
				t.Errorf("skill %s missing description in frontmatter", name)
			}

			// Verify original content is included after frontmatter
			if !strings.Contains(content, "# /"+name+" — ") {
				t.Errorf("skill %s missing original heading", name)
			}
		}
	})

	t.Run("does not overwrite existing native skills", func(t *testing.T) {
		targetDir := filepath.Join(t.TempDir(), ".claude", "skills")
		skillDir := filepath.Join(targetDir, "beans-bplan")
		os.MkdirAll(skillDir, 0755)

		customContent := []byte("---\nname: bplan\ndescription: Custom\n---\n\n# Custom")
		os.WriteFile(filepath.Join(skillDir, "SKILL.md"), customContent, 0644)

		installed, err := installSkills(targetDir, skillFormatNative, false)
		if err != nil {
			t.Fatalf("installSkills() error = %v", err)
		}
		if installed != 4 {
			t.Errorf("installed = %d, want 4 (bplan should be skipped)", installed)
		}

		// Verify custom file was preserved
		data, _ := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
		if string(data) != string(customContent) {
			t.Error("custom SKILL.md was overwritten")
		}
	})

	t.Run("force overwrites existing native skills", func(t *testing.T) {
		targetDir := filepath.Join(t.TempDir(), ".claude", "skills")
		skillDir := filepath.Join(targetDir, "beans-bplan")
		os.MkdirAll(skillDir, 0755)

		os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Custom"), 0644)

		installed, err := installSkills(targetDir, skillFormatNative, true)
		if err != nil {
			t.Fatalf("installSkills() error = %v", err)
		}
		if installed != 5 {
			t.Errorf("installed = %d, want 5", installed)
		}

		data, _ := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
		if string(data) == "# Custom" {
			t.Error("SKILL.md was NOT overwritten with force=true")
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
		if tools[0].Format != skillFormatNative {
			t.Errorf("Claude format = %d, want skillFormatNative", tools[0].Format)
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
	t.Run("native format returns skills root", func(t *testing.T) {
		home := "/home/user"
		tool := agentTool{Name: "Claude", DirName: ".claude", Format: skillFormatNative}

		got := skillsDir(home, tool)
		want := filepath.Join(home, ".claude", "skills")
		if got != want {
			t.Errorf("skillsDir() = %s, want %s", got, want)
		}
	})

	t.Run("flat format returns beans subdirectory", func(t *testing.T) {
		home := "/home/user"
		tool := agentTool{Name: "Codex", DirName: ".codex", Format: skillFormatFlat}

		got := skillsDir(home, tool)
		want := filepath.Join(home, ".codex", "skills", "beans")
		if got != want {
			t.Errorf("skillsDir() = %s, want %s", got, want)
		}
	})
}

func TestInstallSkillsForTools(t *testing.T) {
	t.Run("installs to both tools with correct formats", func(t *testing.T) {
		home := t.TempDir()
		os.MkdirAll(filepath.Join(home, ".claude"), 0755)
		os.MkdirAll(filepath.Join(home, ".codex"), 0755)

		tools := []agentTool{
			{Name: "Claude", DirName: ".claude", Format: skillFormatNative},
			{Name: "Codex", DirName: ".codex", Format: skillFormatFlat},
		}

		err := installSkillsForTools(home, tools, false)
		if err != nil {
			t.Fatalf("installSkillsForTools() error = %v", err)
		}

		// Verify Claude has native format
		for _, name := range []string{"bplan", "brefine", "breview", "bship", "binvestigate"} {
			path := filepath.Join(home, ".claude", "skills", "beans-"+name, "SKILL.md")
			if _, err := os.Stat(path); err != nil {
				t.Errorf("expected Claude skill SKILL.md at %s", path)
			}
		}

		// Verify Codex has flat format
		for _, name := range []string{"bplan.md", "brefine.md", "breview.md", "bship.md", "binvestigate.md"} {
			path := filepath.Join(home, ".codex", "skills", "beans", name)
			if _, err := os.Stat(path); err != nil {
				t.Errorf("expected Codex skill file %s", path)
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

func TestExtractDescriptionFromContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "heading with separator",
			content: "# /bplan — Critical Bean Planning\n\nDetails.",
			want:    "Critical Bean Planning",
		},
		{
			name:    "heading without separator",
			content: "# Code Review Skill\n\nDetails.",
			want:    "Code Review Skill",
		},
		{
			name:    "no heading",
			content: "Just some text without headings.",
			want:    "",
		},
		{
			name:    "empty content",
			content: "",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractDescriptionFromContent(tt.content)
			if got != tt.want {
				t.Errorf("extractDescriptionFromContent() = %q, want %q", got, tt.want)
			}
		})
	}
}
