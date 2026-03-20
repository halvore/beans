package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/hmans/beans/pkg/localregistry"
	"github.com/spf13/cobra"
)

// setupProjectsTest creates a temporary local registry with a registered project.
// Returns the registry, project path, and project slug.
func setupProjectsTest(t *testing.T) (*localregistry.Registry, string, string) {
	t.Helper()
	localDir := t.TempDir()
	t.Setenv(localregistry.EnvLocalDir, localDir)

	projectDir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	reg, err := localregistry.Load()
	if err != nil {
		t.Fatal(err)
	}

	entry, err := reg.Register(projectDir, filepath.Base(projectDir))
	if err != nil {
		t.Fatal(err)
	}
	if err := reg.Save(); err != nil {
		t.Fatal(err)
	}

	return reg, projectDir, entry.Slug
}

func newProjectsRoot() *cobra.Command {
	root := NewRootCmd()
	RegisterProjectsCmd(root)
	return root
}

func TestProjectsList(t *testing.T) {
	t.Run("empty registry", func(t *testing.T) {
		localDir := t.TempDir()
		t.Setenv(localregistry.EnvLocalDir, localDir)

		root := newProjectsRoot()
		root.SetArgs([]string{"projects", "list", "--json"})

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := root.Execute()

		w.Close()
		os.Stdout = old

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var buf bytes.Buffer
		buf.ReadFrom(r)

		var entries []localregistry.ProjectEntry
		if err := json.Unmarshal(buf.Bytes(), &entries); err != nil {
			t.Fatalf("failed to parse JSON: %v (output: %s)", err, buf.String())
		}

		if len(entries) != 0 {
			t.Errorf("expected 0 entries, got %d", len(entries))
		}
	})

	t.Run("lists registered projects as JSON", func(t *testing.T) {
		_, projectDir, _ := setupProjectsTest(t)

		root := newProjectsRoot()
		root.SetArgs([]string{"projects", "list", "--json"})

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := root.Execute()

		w.Close()
		os.Stdout = old

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var buf bytes.Buffer
		buf.ReadFrom(r)

		var entries []localregistry.ProjectEntry
		if err := json.Unmarshal(buf.Bytes(), &entries); err != nil {
			t.Fatalf("failed to parse JSON: %v (output: %s)", err, buf.String())
		}

		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}
		if entries[0].Path != projectDir {
			t.Errorf("expected path %s, got %s", projectDir, entries[0].Path)
		}
	})
}

func TestProjectsRemove(t *testing.T) {
	t.Run("removes registered project", func(t *testing.T) {
		_, projectDir, _ := setupProjectsTest(t)

		root := newProjectsRoot()
		root.SetArgs([]string{"projects", "remove", projectDir})

		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify removed from registry
		reg, err := localregistry.Load()
		if err != nil {
			t.Fatal(err)
		}
		if reg.Lookup(projectDir) != nil {
			t.Error("expected project to be unregistered")
		}
	})

	t.Run("removes project with --delete-data", func(t *testing.T) {
		reg, projectDir, slug := setupProjectsTest(t)

		projDir, err := reg.ProjectDir(slug)
		if err != nil {
			t.Fatal(err)
		}
		// Verify project dir exists before removal
		if _, err := os.Stat(projDir); os.IsNotExist(err) {
			t.Fatal("expected project dir to exist before removal")
		}

		root := newProjectsRoot()
		root.SetArgs([]string{"projects", "remove", "--delete-data", projectDir})

		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify project dir is deleted
		if _, err := os.Stat(projDir); !os.IsNotExist(err) {
			t.Error("expected project dir to be deleted")
		}
	})

	t.Run("error for unregistered project", func(t *testing.T) {
		localDir := t.TempDir()
		t.Setenv(localregistry.EnvLocalDir, localDir)

		root := newProjectsRoot()
		root.SetArgs([]string{"projects", "remove", "/nonexistent/path"})

		err := root.Execute()
		if err == nil {
			t.Fatal("expected error for nonexistent project")
		}
	})
}
