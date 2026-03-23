package commands

import (
	"os"
	"path/filepath"
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
