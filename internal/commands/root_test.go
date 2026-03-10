package commands

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hmans/beans/internal/bean"
	"github.com/hmans/beans/internal/beancore"
	"github.com/hmans/beans/internal/config"
)

func TestResolveBeansPath(t *testing.T) {
	// Create a valid beans directory for tests that need one
	tmpDir := t.TempDir()
	validBeansDir := filepath.Join(tmpDir, ".beans")
	if err := os.MkdirAll(validBeansDir, 0755); err != nil {
		t.Fatalf("failed to create test .beans dir: %v", err)
	}

	altBeansDir := filepath.Join(tmpDir, "alt-beans")
	if err := os.MkdirAll(altBeansDir, 0755); err != nil {
		t.Fatalf("failed to create alt beans dir: %v", err)
	}

	// Config that points to the valid beans dir
	cfg := config.Default()
	cfg.SetConfigDir(tmpDir)

	t.Run("flag takes highest precedence", func(t *testing.T) {
		t.Setenv("BEANS_PATH", altBeansDir)

		got, _, err := resolveBeansPath(validBeansDir, cfg)
		if err != nil {
			t.Fatalf("resolveBeansPath() error = %v", err)
		}
		if got != validBeansDir {
			t.Errorf("expected flag path %q, got %q", validBeansDir, got)
		}
	})

	t.Run("flag overrides env var", func(t *testing.T) {
		t.Setenv("BEANS_PATH", "/nonexistent/should/not/be/used")

		got, _, err := resolveBeansPath(validBeansDir, cfg)
		if err != nil {
			t.Fatalf("resolveBeansPath() error = %v", err)
		}
		if got != validBeansDir {
			t.Errorf("expected flag path %q, got %q", validBeansDir, got)
		}
	})

	t.Run("env var used when flag is empty", func(t *testing.T) {
		t.Setenv("BEANS_PATH", altBeansDir)

		got, _, err := resolveBeansPath("", cfg)
		if err != nil {
			t.Fatalf("resolveBeansPath() error = %v", err)
		}
		if got != altBeansDir {
			t.Errorf("expected env var path %q, got %q", altBeansDir, got)
		}
	})

	t.Run("config used when flag and env var are empty", func(t *testing.T) {
		t.Setenv("BEANS_PATH", "")

		got, _, err := resolveBeansPath("", cfg)
		if err != nil {
			t.Fatalf("resolveBeansPath() error = %v", err)
		}
		expected := cfg.ResolveBeansPath()
		if got != expected {
			t.Errorf("expected config path %q, got %q", expected, got)
		}
	})

	t.Run("invalid flag path returns error", func(t *testing.T) {
		_, _, err := resolveBeansPath("/nonexistent/path", cfg)
		if err == nil {
			t.Fatal("expected error for invalid flag path, got nil")
		}
		if !strings.Contains(err.Error(), "does not exist or is not a directory") {
			t.Errorf("expected 'does not exist' error, got %q", err.Error())
		}
	})

	t.Run("invalid env var path returns error", func(t *testing.T) {
		t.Setenv("BEANS_PATH", "/nonexistent/env/path")

		_, _, err := resolveBeansPath("", cfg)
		if err == nil {
			t.Fatal("expected error for invalid env var path, got nil")
		}
		if !strings.Contains(err.Error(), "does not exist or is not a directory") {
			t.Errorf("expected 'does not exist' error, got %q", err.Error())
		}
	})

	t.Run("invalid config path returns init suggestion", func(t *testing.T) {
		t.Setenv("BEANS_PATH", "")

		// Config pointing to a nonexistent directory
		badCfg := config.Default()
		badCfg.SetConfigDir("/nonexistent/config/dir")

		_, _, err := resolveBeansPath("", badCfg)
		if err == nil {
			t.Fatal("expected error for invalid config path, got nil")
		}
		if !strings.Contains(err.Error(), "beans init") {
			t.Errorf("expected error to suggest 'beans init', got %q", err.Error())
		}
	})

	t.Run("file path rejected as not a directory", func(t *testing.T) {
		// Create a regular file (not a directory)
		filePath := filepath.Join(tmpDir, "not-a-dir")
		if err := os.WriteFile(filePath, []byte("hello"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		_, _, err := resolveBeansPath(filePath, cfg)
		if err == nil {
			t.Fatal("expected error for file path (not directory), got nil")
		}
		if !strings.Contains(err.Error(), "does not exist or is not a directory") {
			t.Errorf("expected 'does not exist' error, got %q", err.Error())
		}
	})

	t.Run("non-worktree returns empty mirror", func(t *testing.T) {
		t.Setenv("BEANS_PATH", "")

		_, mirror, err := resolveBeansPath("", cfg)
		if err != nil {
			t.Fatalf("resolveBeansPath() error = %v", err)
		}
		if mirror != "" {
			t.Errorf("expected empty mirror for non-worktree, got %q", mirror)
		}
	})
}

// initGitRepo creates a temporary git repo with an initial commit and .beans/ directory.
func initGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	commands := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	}
	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v failed: %s: %v", args, out, err)
		}
	}

	// Create .beans/ directory and config file
	beansDir := filepath.Join(dir, ".beans")
	if err := os.MkdirAll(beansDir, 0755); err != nil {
		t.Fatalf("failed to create .beans dir: %v", err)
	}
	// Write a .beans.yml so config.LoadFromDirectory stops walking at the repo root
	if err := os.WriteFile(filepath.Join(dir, ".beans.yml"), []byte(""), 0644); err != nil {
		t.Fatalf("failed to create .beans.yml: %v", err)
	}

	// Initial commit including .beans/
	gitCommands := [][]string{
		{"git", "add", "."},
		{"git", "commit", "--allow-empty", "-m", "initial"},
	}
	for _, args := range gitCommands {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v failed: %s: %v", args, out, err)
		}
	}

	return dir
}

// addGitWorktree creates a secondary worktree for the given repo.
func addGitWorktree(t *testing.T, repoDir string) string {
	t.Helper()
	wtPath := filepath.Join(t.TempDir(), "worktree")
	cmd := exec.Command("git", "worktree", "add", wtPath, "-b", "test-worktree-branch")
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git worktree add failed: %s: %v", out, err)
	}
	return wtPath
}

func TestResolveBeansPath_WorktreeReturnsMirror(t *testing.T) {
	t.Setenv("BEANS_PATH", "")

	repoDir := initGitRepo(t)
	wtDir := addGitWorktree(t, repoDir)

	// Ensure the worktree also has a .beans/ directory (as it would from the commit)
	wtBeansDir := filepath.Join(wtDir, ".beans")
	if err := os.MkdirAll(wtBeansDir, 0755); err != nil {
		t.Fatalf("failed to create worktree .beans dir: %v", err)
	}

	// Load config from worktree (simulates what PersistentPreRunE does)
	wtCfg, err := config.LoadFromDirectory(wtDir)
	if err != nil {
		t.Fatalf("LoadFromDirectory() error = %v", err)
	}

	primary, mirror, err := resolveBeansPath("", wtCfg)
	if err != nil {
		t.Fatalf("resolveBeansPath() error = %v", err)
	}

	// Resolve symlinks for comparison (macOS /tmp -> /private/tmp)
	expectedPrimary, _ := filepath.EvalSymlinks(filepath.Join(repoDir, ".beans"))
	actualPrimary, _ := filepath.EvalSymlinks(primary)
	expectedMirror, _ := filepath.EvalSymlinks(wtBeansDir)
	actualMirror, _ := filepath.EvalSymlinks(mirror)

	if actualPrimary != expectedPrimary {
		t.Errorf("primary = %q, want main repo's .beans/ at %q", actualPrimary, expectedPrimary)
	}
	if actualMirror != expectedMirror {
		t.Errorf("mirror = %q, want worktree's .beans/ at %q", actualMirror, expectedMirror)
	}
}

func TestResolveBeansPath_WorktreeExplicitOverrideSkipsMirror(t *testing.T) {
	repoDir := initGitRepo(t)
	wtDir := addGitWorktree(t, repoDir)

	mainBeansDir := filepath.Join(repoDir, ".beans")

	// With explicit flag, worktree detection is bypassed
	wtCfg, err := config.LoadFromDirectory(wtDir)
	if err != nil {
		t.Fatalf("LoadFromDirectory() error = %v", err)
	}

	primary, mirror, err := resolveBeansPath(mainBeansDir, wtCfg)
	if err != nil {
		t.Fatalf("resolveBeansPath() error = %v", err)
	}

	actualPrimary, _ := filepath.EvalSymlinks(primary)
	expectedPrimary, _ := filepath.EvalSymlinks(mainBeansDir)

	if actualPrimary != expectedPrimary {
		t.Errorf("primary = %q, want %q", actualPrimary, expectedPrimary)
	}
	if mirror != "" {
		t.Errorf("mirror should be empty with explicit flag, got %q", mirror)
	}
}

func TestWorktreeDualWrite_Integration(t *testing.T) {
	t.Setenv("BEANS_PATH", "")

	repoDir := initGitRepo(t)
	wtDir := addGitWorktree(t, repoDir)

	wtBeansDir := filepath.Join(wtDir, ".beans")
	if err := os.MkdirAll(wtBeansDir, 0755); err != nil {
		t.Fatalf("failed to create worktree .beans dir: %v", err)
	}

	// Resolve paths as PersistentPreRunE would
	wtCfg, err := config.LoadFromDirectory(wtDir)
	if err != nil {
		t.Fatalf("LoadFromDirectory() error = %v", err)
	}

	primary, mirror, err := resolveBeansPath("", wtCfg)
	if err != nil {
		t.Fatalf("resolveBeansPath() error = %v", err)
	}

	if mirror == "" {
		t.Fatal("expected non-empty mirror path for worktree")
	}

	// Create core with mirror wired up (same as PersistentPreRunE)
	c := beancore.New(primary, wtCfg)
	c.SetMirrorRoot(mirror)
	c.SetWarnWriter(nil)
	if err := c.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	t.Run("create writes to both", func(t *testing.T) {
		b := &bean.Bean{
			ID:     "wt-1",
			Slug:   "worktree-test",
			Title:  "Worktree Test",
			Status: "todo",
		}
		if err := c.Create(b); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// Primary (main repo) should have the file
		primaryFile := filepath.Join(primary, b.Path)
		if _, err := os.Stat(primaryFile); os.IsNotExist(err) {
			t.Error("bean file missing from primary (main repo) .beans/")
		}

		// Mirror (worktree) should also have the file
		mirrorFile := filepath.Join(mirror, b.Path)
		if _, err := os.Stat(mirrorFile); os.IsNotExist(err) {
			t.Error("bean file missing from mirror (worktree) .beans/")
		}

		// Content should match
		pc, _ := os.ReadFile(primaryFile)
		mc, _ := os.ReadFile(mirrorFile)
		if string(pc) != string(mc) {
			t.Error("primary and mirror file contents differ")
		}
	})

	t.Run("update writes to both", func(t *testing.T) {
		b, err := c.Get("wt-1")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		b.Status = "completed"
		if err := c.Update(b, nil); err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		mirrorFile := filepath.Join(mirror, b.Path)
		mc, err := os.ReadFile(mirrorFile)
		if err != nil {
			t.Fatalf("failed to read mirror file: %v", err)
		}
		if !strings.Contains(string(mc), "status: completed") {
			t.Error("mirror file should contain updated status")
		}
	})

	t.Run("delete removes from both", func(t *testing.T) {
		b := &bean.Bean{
			ID:     "wt-del",
			Slug:   "delete-test",
			Title:  "Delete Test",
			Status: "todo",
		}
		if err := c.Create(b); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		relPath := b.Path

		if err := c.Delete("wt-del"); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		if _, err := os.Stat(filepath.Join(primary, relPath)); !os.IsNotExist(err) {
			t.Error("primary file should be removed after delete")
		}
		if _, err := os.Stat(filepath.Join(mirror, relPath)); !os.IsNotExist(err) {
			t.Error("mirror file should be removed after delete")
		}
	})

	t.Run("archive moves in both", func(t *testing.T) {
		b := &bean.Bean{
			ID:     "wt-arc",
			Slug:   "archive-test",
			Title:  "Archive Test",
			Status: "completed",
		}
		if err := c.Create(b); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		origPath := b.Path

		if err := c.Archive("wt-arc"); err != nil {
			t.Fatalf("Archive() error = %v", err)
		}

		// Original gone from both
		if _, err := os.Stat(filepath.Join(primary, origPath)); !os.IsNotExist(err) {
			t.Error("original primary file should be gone after archive")
		}
		if _, err := os.Stat(filepath.Join(mirror, origPath)); !os.IsNotExist(err) {
			t.Error("original mirror file should be gone after archive")
		}

		// Archived file present in both
		archivedName := filepath.Join("archive", filepath.Base(origPath))
		if _, err := os.Stat(filepath.Join(primary, archivedName)); os.IsNotExist(err) {
			t.Error("archived file missing from primary")
		}
		if _, err := os.Stat(filepath.Join(mirror, archivedName)); os.IsNotExist(err) {
			t.Error("archived file missing from mirror")
		}
	})

	t.Run("reads only come from primary", func(t *testing.T) {
		// Write a bean only to mirror (not through Core)
		fakeContent := []byte("---\ntitle: Ghost\nstatus: todo\ntype: task\n---\n")
		if err := os.WriteFile(filepath.Join(mirror, "ghost--ghost.md"), fakeContent, 0644); err != nil {
			t.Fatalf("failed to write fake mirror bean: %v", err)
		}

		// Core should NOT see it (reads from primary only)
		_, err := c.Get("ghost")
		if err != beancore.ErrNotFound {
			t.Errorf("expected ErrNotFound for mirror-only bean, got %v", err)
		}
	})
}
