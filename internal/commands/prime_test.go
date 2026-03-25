package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/hmans/beans/pkg/config"
	"github.com/hmans/beans/pkg/localregistry"
)

func TestPrimeWithLocalStorage(t *testing.T) {
	localDir := t.TempDir()
	t.Setenv(localregistry.EnvLocalDir, localDir)

	// Resolve symlinks (macOS /var → /private/var) to match os.Getwd().
	projectDir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	// Register the project in the local registry.
	reg := &localregistry.Registry{}
	entry, err := reg.Register(projectDir, "test-project", "")
	if err != nil {
		t.Fatalf("failed to register project: %v", err)
	}
	if err := reg.Save(); err != nil {
		t.Fatalf("failed to save registry: %v", err)
	}

	// Create config and .beans dir in the local project directory.
	localProjectDir, err := reg.ProjectDir(entry.Slug)
	if err != nil {
		t.Fatalf("failed to get project dir: %v", err)
	}
	cfgToSave := config.DefaultWithPrefix("test-project-")
	cfgToSave.SetConfigDir(localProjectDir)
	if err := cfgToSave.Save(localProjectDir); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
	beansDir := filepath.Join(localProjectDir, ".beans")
	if err := os.MkdirAll(beansDir, 0755); err != nil {
		t.Fatalf("failed to create .beans dir: %v", err)
	}

	// Install a skill to $HOME/.claude/skills/ (where local project skills go)
	// so we can verify the path in the output.
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	localSkillsDir := filepath.Join(home, ".claude", "skills")
	if err := os.MkdirAll(localSkillsDir, 0755); err != nil {
		t.Fatalf("failed to create skills dir: %v", err)
	}
	skillFile := filepath.Join(localSkillsDir, "bplan.md")
	// Only write if not already present (avoid clobbering real skills)
	if _, statErr := os.Stat(skillFile); statErr != nil {
		if err := os.WriteFile(skillFile, []byte("# /bplan — Critical Bean Planning\n\nDetails."), 0644); err != nil {
			t.Fatalf("failed to write skill file: %v", err)
		}
		t.Cleanup(func() { os.Remove(skillFile) })
	}

	// Save and restore working directory.
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	if err := os.Chdir(projectDir); err != nil {
		t.Fatal(err)
	}

	// Reset package-level flags to simulate no explicit paths.
	origBeansPath := beansPath
	origConfigPath := configPath
	beansPath = ""
	configPath = ""
	t.Cleanup(func() {
		beansPath = origBeansPath
		configPath = origConfigPath
	})

	t.Run("produces output for local storage project", func(t *testing.T) {
		// Capture stdout.
		var buf bytes.Buffer
		rootCmd := NewRootCmd()
		RegisterPrimeCmd(rootCmd)
		rootCmd.SetOut(&buf)
		rootCmd.SetArgs([]string{"prime"})

		// prime writes directly to os.Stdout, so redirect it.
		oldStdout := os.Stdout
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		os.Stdout = w

		if err := rootCmd.Execute(); err != nil {
			w.Close()
			os.Stdout = oldStdout
			t.Fatalf("prime command failed: %v", err)
		}

		w.Close()
		os.Stdout = oldStdout

		var captured bytes.Buffer
		captured.ReadFrom(r)
		output := captured.String()

		if output == "" {
			t.Fatal("expected prime to produce output for local storage project, got empty string")
		}
		if !bytes.Contains([]byte(output), []byte("Beans Usage Guide")) {
			t.Errorf("expected output to contain 'Beans Usage Guide', got:\n%s", output[:min(len(output), 200)])
		}

		// The skills section should reference the $HOME/.claude/skills/ path for local projects.
		if !bytes.Contains([]byte(output), []byte(localSkillsDir)) {
			t.Errorf("expected output to contain skills path %q, got:\n%s", localSkillsDir, output[:min(len(output), 500)])
		}
	})
}

func TestPrimeWithInRepoConfig(t *testing.T) {
	// Resolve symlinks (macOS /var → /private/var).
	projectDir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	// Create an in-repo .beans.yml config.
	cfgToSave := config.DefaultWithPrefix("test-")
	cfgToSave.SetConfigDir(projectDir)
	if err := cfgToSave.Save(projectDir); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Save and restore working directory.
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	if err := os.Chdir(projectDir); err != nil {
		t.Fatal(err)
	}

	origBeansPath := beansPath
	origConfigPath := configPath
	beansPath = ""
	configPath = ""
	t.Cleanup(func() {
		beansPath = origBeansPath
		configPath = origConfigPath
	})

	t.Run("produces output for in-repo project", func(t *testing.T) {
		rootCmd := NewRootCmd()
		RegisterPrimeCmd(rootCmd)
		rootCmd.SetArgs([]string{"prime"})

		oldStdout := os.Stdout
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		os.Stdout = w

		if err := rootCmd.Execute(); err != nil {
			w.Close()
			os.Stdout = oldStdout
			t.Fatalf("prime command failed: %v", err)
		}

		w.Close()
		os.Stdout = oldStdout

		var captured bytes.Buffer
		captured.ReadFrom(r)
		output := captured.String()

		if output == "" {
			t.Fatal("expected prime to produce output for in-repo project, got empty string")
		}
	})
}

func TestDiscoverSkills(t *testing.T) {
	t.Run("discovers flat skill files", func(t *testing.T) {
		dir := t.TempDir()

		// Write test skill files in flat format
		os.WriteFile(filepath.Join(dir, "bplan.md"), []byte("# /bplan — Critical Bean Planning\n\nDetails here."), 0644)
		os.WriteFile(filepath.Join(dir, "breview.md"), []byte("# /breview — Pre-PR Code Review\n\nDetails here."), 0644)
		// Non-md files should be ignored
		os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("not a skill"), 0644)

		skills := discoverSkills(dir)

		if len(skills) != 2 {
			t.Fatalf("expected 2 skills, got %d", len(skills))
		}

		if skills[0].Name != "bplan" {
			t.Errorf("skills[0].Name = %q, want \"bplan\"", skills[0].Name)
		}
		if skills[0].Description != "Critical Bean Planning" {
			t.Errorf("skills[0].Description = %q, want \"Critical Bean Planning\"", skills[0].Description)
		}
		if skills[1].Name != "breview" {
			t.Errorf("skills[1].Name = %q, want \"breview\"", skills[1].Name)
		}
	})

	t.Run("discovers native format skills", func(t *testing.T) {
		dir := t.TempDir()

		// Create native format: beans-<name>/SKILL.md
		os.MkdirAll(filepath.Join(dir, "beans-bplan"), 0755)
		os.WriteFile(filepath.Join(dir, "beans-bplan", "SKILL.md"),
			[]byte("---\nname: bplan\ndescription: Critical Bean Planning\n---\n\n# /bplan — Critical Bean Planning\n\nDetails."), 0644)

		os.MkdirAll(filepath.Join(dir, "beans-breview"), 0755)
		os.WriteFile(filepath.Join(dir, "beans-breview", "SKILL.md"),
			[]byte("---\nname: breview\ndescription: Pre-PR Code Review\n---\n\n# /breview — Pre-PR Code Review\n\nDetails."), 0644)

		// Non-beans directories should be ignored
		os.MkdirAll(filepath.Join(dir, "other-skill"), 0755)
		os.WriteFile(filepath.Join(dir, "other-skill", "SKILL.md"), []byte("# Other"), 0644)

		skills := discoverSkills(dir)

		if len(skills) != 2 {
			t.Fatalf("expected 2 skills, got %d", len(skills))
		}

		if skills[0].Name != "bplan" {
			t.Errorf("skills[0].Name = %q, want \"bplan\"", skills[0].Name)
		}
		if skills[0].Description != "Critical Bean Planning" {
			t.Errorf("skills[0].Description = %q, want \"Critical Bean Planning\"", skills[0].Description)
		}
		if skills[1].Name != "breview" {
			t.Errorf("skills[1].Name = %q, want \"breview\"", skills[1].Name)
		}
	})

	t.Run("returns nil for missing directory", func(t *testing.T) {
		skills := discoverSkills("/nonexistent/path")
		if skills != nil {
			t.Errorf("expected nil, got %v", skills)
		}
	})

	t.Run("returns nil for empty directory", func(t *testing.T) {
		dir := t.TempDir()
		skills := discoverSkills(dir)
		if skills != nil {
			t.Errorf("expected nil, got %v", skills)
		}
	})
}

func TestExtractSkillDescription(t *testing.T) {
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
			name:    "empty file",
			content: "",
			want:    "",
		},
		{
			name:    "with YAML frontmatter and heading",
			content: "---\nname: bplan\ndescription: From Frontmatter\n---\n\n# /bplan — Critical Bean Planning\n\nDetails.",
			want:    "Critical Bean Planning",
		},
		{
			name:    "with YAML frontmatter only",
			content: "---\nname: bplan\ndescription: From Frontmatter\n---\n\nNo heading here.",
			want:    "From Frontmatter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "skill.md")
			os.WriteFile(path, []byte(tt.content), 0644)

			got := extractSkillDescription(path)
			if got != tt.want {
				t.Errorf("extractSkillDescription() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrimeWithNoProject(t *testing.T) {
	// Use a temp dir with no local registry and no config.
	localDir := t.TempDir()
	t.Setenv(localregistry.EnvLocalDir, localDir)

	projectDir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	if err := os.Chdir(projectDir); err != nil {
		t.Fatal(err)
	}

	origBeansPath := beansPath
	origConfigPath := configPath
	beansPath = ""
	configPath = ""
	t.Cleanup(func() {
		beansPath = origBeansPath
		configPath = origConfigPath
	})

	t.Run("outputs initialization prompt when no project exists", func(t *testing.T) {
		rootCmd := NewRootCmd()
		RegisterPrimeCmd(rootCmd)
		rootCmd.SetArgs([]string{"prime"})

		oldStdout := os.Stdout
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		os.Stdout = w

		if err := rootCmd.Execute(); err != nil {
			w.Close()
			os.Stdout = oldStdout
			t.Fatalf("prime command should not error: %v", err)
		}

		w.Close()
		os.Stdout = oldStdout

		var captured bytes.Buffer
		captured.ReadFrom(r)
		output := captured.String()

		if output == "" {
			t.Fatal("expected initialization prompt when no project exists, got empty string")
		}
		if !bytes.Contains([]byte(output), []byte("Beans Is Not Initialized")) {
			t.Errorf("expected output to contain 'Beans Is Not Initialized', got:\n%s", output[:min(len(output), 200)])
		}
		if !bytes.Contains([]byte(output), []byte("beans init --local")) {
			t.Errorf("expected output to mention 'beans init --local', got:\n%s", output[:min(len(output), 200)])
		}
		if !bytes.Contains([]byte(output), []byte("MUST ask the user")) {
			t.Errorf("expected output to instruct agent to ask the user, got:\n%s", output[:min(len(output), 200)])
		}
	})
}
