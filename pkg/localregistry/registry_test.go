package localregistry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEmpty(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv(EnvLocalDir, tmp)

	reg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(reg.Projects) != 0 {
		t.Fatalf("expected empty registry, got %d projects", len(reg.Projects))
	}
}

func TestRegisterAndLookup(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv(EnvLocalDir, tmp)

	reg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	projectPath := filepath.Join(tmp, "myproject")
	if err := os.MkdirAll(projectPath, 0o755); err != nil {
		t.Fatal(err)
	}

	entry, err := reg.Register(projectPath, "My Project")
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	if entry.Slug != "my-project" {
		t.Errorf("slug = %q, want %q", entry.Slug, "my-project")
	}
	if entry.Path != projectPath {
		t.Errorf("path = %q, want %q", entry.Path, projectPath)
	}
	if entry.RegisteredAt.IsZero() {
		t.Error("registered_at is zero")
	}

	// Lookup by path.
	found := reg.Lookup(projectPath)
	if found == nil {
		t.Fatal("Lookup() returned nil")
	}
	if found.Slug != "my-project" {
		t.Errorf("lookup slug = %q, want %q", found.Slug, "my-project")
	}

	// Verify project directory was created.
	beansDir, _ := reg.ProjectBeansDir(entry.Slug)
	if _, err := os.Stat(beansDir); os.IsNotExist(err) {
		t.Errorf("beans directory not created at %s", beansDir)
	}
}

func TestRegisterIdempotent(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv(EnvLocalDir, tmp)

	reg, _ := Load()

	projectPath := filepath.Join(tmp, "proj")
	os.MkdirAll(projectPath, 0o755)

	entry1, _ := reg.Register(projectPath, "proj")
	entry2, _ := reg.Register(projectPath, "proj")

	if entry1.Slug != entry2.Slug {
		t.Errorf("second register changed slug: %q -> %q", entry1.Slug, entry2.Slug)
	}
	if len(reg.Projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(reg.Projects))
	}
}

func TestRegisterSlugCollision(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv(EnvLocalDir, tmp)

	reg, _ := Load()

	path1 := filepath.Join(tmp, "a", "myapp")
	path2 := filepath.Join(tmp, "b", "myapp")
	os.MkdirAll(path1, 0o755)
	os.MkdirAll(path2, 0o755)

	e1, err := reg.Register(path1, "myapp")
	if err != nil {
		t.Fatal(err)
	}
	e2, err := reg.Register(path2, "myapp")
	if err != nil {
		t.Fatal(err)
	}

	if e1.Slug == e2.Slug {
		t.Error("collision not resolved: both slugs are the same")
	}
	if e1.Slug != "myapp" {
		t.Errorf("first slug should be 'myapp', got %q", e1.Slug)
	}
	// Second slug should have hash suffix.
	if len(e2.Slug) <= len("myapp") {
		t.Errorf("second slug should have suffix, got %q", e2.Slug)
	}
}

func TestRegisterFallbackToBasename(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv(EnvLocalDir, tmp)

	reg, _ := Load()
	projectPath := filepath.Join(tmp, "MyApp")
	os.MkdirAll(projectPath, 0o755)

	entry, err := reg.Register(projectPath, "")
	if err != nil {
		t.Fatal(err)
	}
	if entry.Slug != "myapp" {
		t.Errorf("slug = %q, want %q", entry.Slug, "myapp")
	}
}

func TestUnregister(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv(EnvLocalDir, tmp)

	reg, _ := Load()
	projectPath := filepath.Join(tmp, "proj")
	os.MkdirAll(projectPath, 0o755)

	reg.Register(projectPath, "proj")

	if !reg.Unregister(projectPath) {
		t.Error("Unregister() returned false for existing project")
	}
	if len(reg.Projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(reg.Projects))
	}

	if reg.Unregister(projectPath) {
		t.Error("Unregister() returned true for non-existent project")
	}
}

func TestLookupNotFound(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv(EnvLocalDir, tmp)

	reg, _ := Load()

	if reg.Lookup("/nonexistent") != nil {
		t.Error("expected nil for unknown path")
	}
}

func TestSaveAndReload(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv(EnvLocalDir, tmp)

	reg, _ := Load()

	projectPath := filepath.Join(tmp, "proj")
	os.MkdirAll(projectPath, 0o755)

	reg.Register(projectPath, "My Project")
	if err := reg.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Reload.
	reg2, err := Load()
	if err != nil {
		t.Fatalf("Load() after save error: %v", err)
	}
	if len(reg2.Projects) != 1 {
		t.Fatalf("expected 1 project after reload, got %d", len(reg2.Projects))
	}
	if reg2.Projects[0].Slug != "my-project" {
		t.Errorf("slug after reload = %q, want %q", reg2.Projects[0].Slug, "my-project")
	}
	if reg2.Projects[0].Path != projectPath {
		t.Errorf("path after reload = %q, want %q", reg2.Projects[0].Path, projectPath)
	}
}

func TestLocalDirEnvOverride(t *testing.T) {
	t.Setenv(EnvLocalDir, "/custom/path")

	dir, err := LocalDir()
	if err != nil {
		t.Fatal(err)
	}
	if dir != "/custom/path" {
		t.Errorf("LocalDir() = %q, want %q", dir, "/custom/path")
	}
}

func TestProjectDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv(EnvLocalDir, tmp)

	reg := &Registry{}
	dir, err := reg.ProjectDir("myapp")
	if err != nil {
		t.Fatal(err)
	}
	expected := filepath.Join(tmp, "projects", "myapp")
	if dir != expected {
		t.Errorf("ProjectDir() = %q, want %q", dir, expected)
	}
}

func TestProjectBeansDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv(EnvLocalDir, tmp)

	reg := &Registry{}
	dir, err := reg.ProjectBeansDir("myapp")
	if err != nil {
		t.Fatal(err)
	}
	expected := filepath.Join(tmp, "projects", "myapp", ".beans")
	if dir != expected {
		t.Errorf("ProjectBeansDir() = %q, want %q", dir, expected)
	}
}
