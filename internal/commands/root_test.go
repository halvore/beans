package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hmans/beans/pkg/config"
	"github.com/hmans/beans/pkg/localregistry"
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

		got, err := resolveBeansPath(validBeansDir, cfg)
		if err != nil {
			t.Fatalf("resolveBeansPath() error = %v", err)
		}
		if got != validBeansDir {
			t.Errorf("expected flag path %q, got %q", validBeansDir, got)
		}
	})

	t.Run("flag overrides env var", func(t *testing.T) {
		t.Setenv("BEANS_PATH", "/nonexistent/should/not/be/used")

		got, err := resolveBeansPath(validBeansDir, cfg)
		if err != nil {
			t.Fatalf("resolveBeansPath() error = %v", err)
		}
		if got != validBeansDir {
			t.Errorf("expected flag path %q, got %q", validBeansDir, got)
		}
	})

	t.Run("env var used when flag is empty", func(t *testing.T) {
		t.Setenv("BEANS_PATH", altBeansDir)

		got, err := resolveBeansPath("", cfg)
		if err != nil {
			t.Fatalf("resolveBeansPath() error = %v", err)
		}
		if got != altBeansDir {
			t.Errorf("expected env var path %q, got %q", altBeansDir, got)
		}
	})

	t.Run("config used when flag and env var are empty", func(t *testing.T) {
		t.Setenv("BEANS_PATH", "")

		got, err := resolveBeansPath("", cfg)
		if err != nil {
			t.Fatalf("resolveBeansPath() error = %v", err)
		}
		expected := cfg.ResolveBeansPath()
		if got != expected {
			t.Errorf("expected config path %q, got %q", expected, got)
		}
	})

	t.Run("invalid flag path returns error", func(t *testing.T) {
		_, err := resolveBeansPath("/nonexistent/path", cfg)
		if err == nil {
			t.Fatal("expected error for invalid flag path, got nil")
		}
		if !strings.Contains(err.Error(), "does not exist or is not a directory") {
			t.Errorf("expected 'does not exist' error, got %q", err.Error())
		}
	})

	t.Run("invalid env var path returns error", func(t *testing.T) {
		t.Setenv("BEANS_PATH", "/nonexistent/env/path")

		_, err := resolveBeansPath("", cfg)
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

		_, err := resolveBeansPath("", badCfg)
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

		_, err := resolveBeansPath(filePath, cfg)
		if err == nil {
			t.Fatal("expected error for file path (not directory), got nil")
		}
		if !strings.Contains(err.Error(), "does not exist or is not a directory") {
			t.Errorf("expected 'does not exist' error, got %q", err.Error())
		}
	})
}

func TestLoadFromLocalRegistry(t *testing.T) {
	t.Run("returns default config when registry does not exist", func(t *testing.T) {
		localDir := t.TempDir()
		t.Setenv(localregistry.EnvLocalDir, localDir)
		// No registry.yml exists

		projectDir := t.TempDir()
		cfg, err := loadFromLocalRegistry(projectDir)
		if err != nil {
			t.Fatalf("loadFromLocalRegistry() error = %v", err)
		}
		if cfg.ConfigDir() != projectDir {
			t.Errorf("expected configDir=%q, got %q", projectDir, cfg.ConfigDir())
		}
	})

	t.Run("returns default config when project not in registry", func(t *testing.T) {
		localDir := t.TempDir()
		t.Setenv(localregistry.EnvLocalDir, localDir)

		// Create an empty registry
		reg := &localregistry.Registry{}
		if err := reg.Save(); err != nil {
			t.Fatalf("failed to save registry: %v", err)
		}

		projectDir := t.TempDir()
		cfg, err := loadFromLocalRegistry(projectDir)
		if err != nil {
			t.Fatalf("loadFromLocalRegistry() error = %v", err)
		}
		if cfg.ConfigDir() != projectDir {
			t.Errorf("expected configDir=%q, got %q", projectDir, cfg.ConfigDir())
		}
	})

	t.Run("loads config from local registry when project is registered", func(t *testing.T) {
		localDir := t.TempDir()
		t.Setenv(localregistry.EnvLocalDir, localDir)

		projectDir := t.TempDir()

		// Register the project
		reg := &localregistry.Registry{}
		entry, err := reg.Register(projectDir, "test-project")
		if err != nil {
			t.Fatalf("failed to register project: %v", err)
		}
		if err := reg.Save(); err != nil {
			t.Fatalf("failed to save registry: %v", err)
		}

		// Create a config file in the local project directory
		localProjectDir, err := reg.ProjectDir(entry.Slug)
		if err != nil {
			t.Fatalf("failed to get project dir: %v", err)
		}
		cfgToSave := config.DefaultWithPrefix("test-project-")
		cfgToSave.Project.Name = "test-project"
		cfgToSave.SetConfigDir(localProjectDir)
		if err := cfgToSave.Save(localProjectDir); err != nil {
			t.Fatalf("failed to save config: %v", err)
		}

		// Load from local registry
		cfg, err := loadFromLocalRegistry(projectDir)
		if err != nil {
			t.Fatalf("loadFromLocalRegistry() error = %v", err)
		}

		// Config should be loaded from the local project directory
		if cfg.ConfigDir() != localProjectDir {
			t.Errorf("expected configDir=%q, got %q", localProjectDir, cfg.ConfigDir())
		}
		if cfg.GetProjectName() != "test-project" {
			t.Errorf("expected project name %q, got %q", "test-project", cfg.GetProjectName())
		}
		if cfg.Beans.Prefix != "test-project-" {
			t.Errorf("expected prefix %q, got %q", "test-project-", cfg.Beans.Prefix)
		}
	})

	t.Run("sets project root to actual project path", func(t *testing.T) {
		localDir := t.TempDir()
		t.Setenv(localregistry.EnvLocalDir, localDir)

		projectDir := t.TempDir()

		// Register the project
		reg := &localregistry.Registry{}
		entry, err := reg.Register(projectDir, "test-project")
		if err != nil {
			t.Fatalf("failed to register project: %v", err)
		}
		if err := reg.Save(); err != nil {
			t.Fatalf("failed to save registry: %v", err)
		}

		// Create a config file in the local project directory
		localProjectDir, err := reg.ProjectDir(entry.Slug)
		if err != nil {
			t.Fatalf("failed to get project dir: %v", err)
		}
		cfgToSave := config.DefaultWithPrefix("test-project-")
		cfgToSave.SetConfigDir(localProjectDir)
		if err := cfgToSave.Save(localProjectDir); err != nil {
			t.Fatalf("failed to save config: %v", err)
		}

		// Load from local registry
		cfg, err := loadFromLocalRegistry(projectDir)
		if err != nil {
			t.Fatalf("loadFromLocalRegistry() error = %v", err)
		}

		// ProjectRoot should be the actual project directory, not the local registry dir
		if cfg.ProjectRoot() != projectDir {
			t.Errorf("expected ProjectRoot=%q, got %q", projectDir, cfg.ProjectRoot())
		}
		// ConfigDir should still be the local project directory
		if cfg.ConfigDir() != localProjectDir {
			t.Errorf("expected ConfigDir=%q, got %q", localProjectDir, cfg.ConfigDir())
		}
	})

	t.Run("resolves beans path from local registry config", func(t *testing.T) {
		localDir := t.TempDir()
		t.Setenv(localregistry.EnvLocalDir, localDir)

		projectDir := t.TempDir()

		// Register the project
		reg := &localregistry.Registry{}
		entry, err := reg.Register(projectDir, "test-project")
		if err != nil {
			t.Fatalf("failed to register project: %v", err)
		}
		if err := reg.Save(); err != nil {
			t.Fatalf("failed to save registry: %v", err)
		}

		// Create config in local project dir
		localProjectDir, err := reg.ProjectDir(entry.Slug)
		if err != nil {
			t.Fatalf("failed to get project dir: %v", err)
		}
		cfgToSave := config.DefaultWithPrefix("test-project-")
		cfgToSave.SetConfigDir(localProjectDir)
		if err := cfgToSave.Save(localProjectDir); err != nil {
			t.Fatalf("failed to save config: %v", err)
		}

		// Load config and resolve beans path
		cfg, err := loadFromLocalRegistry(projectDir)
		if err != nil {
			t.Fatalf("loadFromLocalRegistry() error = %v", err)
		}

		beansDir := cfg.ResolveBeansPath()
		expectedBeansDir := filepath.Join(localProjectDir, ".beans")
		if beansDir != expectedBeansDir {
			t.Errorf("expected beans path %q, got %q", expectedBeansDir, beansDir)
		}
	})
}
