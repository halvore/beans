package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hmans/beans/pkg/localregistry"
)

func TestInitLocal(t *testing.T) {
	// Set up a temporary local registry directory so we don't touch real state.
	localDir := t.TempDir()
	t.Setenv(localregistry.EnvLocalDir, localDir)

	// Create a fake project directory.
	// Resolve symlinks (macOS /var → /private/var) to match os.Getwd().
	projectDir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
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

	t.Run("creates local registry entry and config", func(t *testing.T) {
		// Reset flags for the test.
		initLocal = true
		initJSON = false

		err := initLocalProject()
		if err != nil {
			t.Fatalf("initLocalProject() error = %v", err)
		}

		// Verify registry was created and has an entry.
		reg, err := localregistry.Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		entry := reg.Lookup(projectDir)
		if entry == nil {
			t.Fatal("expected project to be registered")
		}

		// Verify the .beans directory was created inside the local project dir.
		beansDir, err := reg.ProjectBeansDir(entry.Slug)
		if err != nil {
			t.Fatalf("ProjectBeansDir() error = %v", err)
		}
		if _, err := os.Stat(beansDir); os.IsNotExist(err) {
			t.Errorf("expected .beans dir at %s", beansDir)
		}

		// Verify .gitignore was created inside the beans dir.
		gitignore := filepath.Join(beansDir, ".gitignore")
		if _, err := os.Stat(gitignore); os.IsNotExist(err) {
			t.Errorf("expected .gitignore at %s", gitignore)
		}

		// Verify config was saved in the local project dir.
		projDir, err := reg.ProjectDir(entry.Slug)
		if err != nil {
			t.Fatalf("ProjectDir() error = %v", err)
		}
		configPath := filepath.Join(projDir, ".beans.yml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Errorf("expected config at %s", configPath)
		}
	})

	t.Run("does not modify project directory", func(t *testing.T) {
		entries, err := os.ReadDir(projectDir)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			if e.Name() == ".beans" || e.Name() == ".beans.yml" {
				t.Errorf("project directory should not contain %s", e.Name())
			}
		}
	})

	t.Run("idempotent on second call", func(t *testing.T) {
		initLocal = true
		initJSON = false

		err := initLocalProject()
		if err != nil {
			t.Fatalf("second initLocalProject() error = %v", err)
		}

		reg, err := localregistry.Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		// Should still have exactly one entry for this project.
		count := 0
		for _, p := range reg.Projects {
			if p.Path == projectDir {
				count++
			}
		}
		if count != 1 {
			t.Errorf("expected 1 registry entry, got %d", count)
		}
	})
}
