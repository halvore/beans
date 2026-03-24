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

	entry, err := reg.Register(projectPath, "My Project", "")
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

	entry1, _ := reg.Register(projectPath, "proj", "")
	entry2, _ := reg.Register(projectPath, "proj", "")

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

	e1, err := reg.Register(path1, "myapp", "")
	if err != nil {
		t.Fatal(err)
	}
	e2, err := reg.Register(path2, "myapp", "")
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

	entry, err := reg.Register(projectPath, "", "")
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

	reg.Register(projectPath, "proj", "")

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

	reg.Register(projectPath, "My Project", "")
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

func TestSlugFromRemoteURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{"HTTPS with .git", "https://github.com/halvore/beans.git", "halvore-beans"},
		{"HTTPS without .git", "https://github.com/halvore/beans", "halvore-beans"},
		{"SSH format", "git@github.com:halvore/beans.git", "halvore-beans"},
		{"SSH without .git", "git@github.com:halvore/beans", "halvore-beans"},
		{"GitLab nested group", "https://gitlab.com/org/subgroup/repo.git", "subgroup-repo"},
		{"SSH nested group", "git@gitlab.com:org/subgroup/repo.git", "subgroup-repo"},
		{"empty string", "", ""},
		{"not a URL", "just-a-name", ""},
		{"single path segment HTTPS", "https://example.com/repo.git", "repo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slugFromRemoteURL(tt.url)
			if got != tt.expected {
				t.Errorf("slugFromRemoteURL(%q) = %q, want %q", tt.url, got, tt.expected)
			}
		})
	}
}

func TestRegisterWithRemoteURL(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv(EnvLocalDir, tmp)

	reg, _ := Load()

	projectPath := filepath.Join(tmp, "myproject")
	os.MkdirAll(projectPath, 0o755)

	entry, err := reg.Register(projectPath, "myproject", "https://github.com/halvore/beans.git")
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	if entry.Slug != "halvore-beans" {
		t.Errorf("slug = %q, want %q", entry.Slug, "halvore-beans")
	}
	if entry.RemoteURL != "https://github.com/halvore/beans.git" {
		t.Errorf("remote_url = %q, want %q", entry.RemoteURL, "https://github.com/halvore/beans.git")
	}
}

func TestRegisterRemoteURLCollision(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv(EnvLocalDir, tmp)

	reg, _ := Load()

	path1 := filepath.Join(tmp, "a", "beans")
	path2 := filepath.Join(tmp, "b", "beans")
	os.MkdirAll(path1, 0o755)
	os.MkdirAll(path2, 0o755)

	// Same owner/repo on different hosts — same slug base.
	e1, err := reg.Register(path1, "", "https://github.com/halvore/beans.git")
	if err != nil {
		t.Fatal(err)
	}
	e2, err := reg.Register(path2, "", "https://gitlab.com/halvore/beans.git")
	if err != nil {
		t.Fatal(err)
	}

	if e1.Slug == e2.Slug {
		t.Error("collision not resolved: both slugs are the same")
	}
	if e1.Slug != "halvore-beans" {
		t.Errorf("first slug should be 'halvore-beans', got %q", e1.Slug)
	}
}

func TestLookupByRemoteURL(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv(EnvLocalDir, tmp)

	reg, _ := Load()

	projectPath := filepath.Join(tmp, "myproject")
	os.MkdirAll(projectPath, 0o755)

	remoteURL := "https://github.com/halvore/beans.git"
	reg.Register(projectPath, "myproject", remoteURL)

	// Lookup by remote URL should find the project.
	found := reg.LookupByRemoteURL(remoteURL)
	if found == nil {
		t.Fatal("LookupByRemoteURL() returned nil for registered remote URL")
	}
	if found.Path != projectPath {
		t.Errorf("path = %q, want %q", found.Path, projectPath)
	}

	// Lookup with different URL should return nil.
	if reg.LookupByRemoteURL("https://github.com/other/repo.git") != nil {
		t.Error("expected nil for unknown remote URL")
	}

	// Lookup with empty URL should return nil.
	if reg.LookupByRemoteURL("") != nil {
		t.Error("expected nil for empty remote URL")
	}
}

func TestLookupByRemoteURLSkipsEmptyEntries(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv(EnvLocalDir, tmp)

	reg, _ := Load()

	// Register a project without a remote URL.
	projectPath := filepath.Join(tmp, "noremote")
	os.MkdirAll(projectPath, 0o755)
	reg.Register(projectPath, "noremote", "")

	// Should not match even if we search for an empty string.
	if reg.LookupByRemoteURL("") != nil {
		t.Error("expected nil when searching with empty remote URL")
	}
}

func TestRegisterFallsBackWithoutRemoteURL(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv(EnvLocalDir, tmp)

	reg, _ := Load()

	projectPath := filepath.Join(tmp, "myapp")
	os.MkdirAll(projectPath, 0o755)

	entry, err := reg.Register(projectPath, "myapp", "")
	if err != nil {
		t.Fatal(err)
	}
	if entry.Slug != "myapp" {
		t.Errorf("slug = %q, want %q", entry.Slug, "myapp")
	}
	if entry.RemoteURL != "" {
		t.Errorf("remote_url should be empty, got %q", entry.RemoteURL)
	}
}
