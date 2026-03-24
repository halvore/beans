package localregistry

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/hmans/beans/pkg/bean"
	"gopkg.in/yaml.v3"
)

const (
	// DefaultLocalDir is the default root for local beans storage.
	DefaultLocalDir = "~/.local/beans"
	// EnvLocalDir overrides the default local directory.
	EnvLocalDir = "BEANS_LOCAL_DIR"
	// RegistryFileName is the name of the registry file.
	RegistryFileName = "registry.yml"
	// ProjectsDir is the subdirectory containing per-project beans.
	ProjectsDir = "projects"
)

// Registry represents the local beans registry.
type Registry struct {
	// Projects is the list of registered projects.
	Projects []ProjectEntry `yaml:"projects"`
}

// ProjectEntry is a single project in the registry.
type ProjectEntry struct {
	// Path is the absolute path to the project root directory.
	Path string `yaml:"path"`
	// Slug is the directory name under projects/.
	Slug string `yaml:"slug"`
	// RemoteURL is the git remote URL (typically origin) used to identify the project.
	// May be empty for projects without a git remote.
	RemoteURL string `yaml:"remote_url,omitempty"`
	// RegisteredAt is the timestamp when the project was registered.
	RegisteredAt time.Time `yaml:"registered_at"`
}

// LocalDir returns the resolved local beans root directory.
// Checks BEANS_LOCAL_DIR env var first, then falls back to DefaultLocalDir.
func LocalDir() (string, error) {
	if dir := os.Getenv(EnvLocalDir); dir != "" {
		return expandHome(dir)
	}
	return expandHome(DefaultLocalDir)
}

// RegistryPath returns the full path to the registry file.
func RegistryPath() (string, error) {
	dir, err := LocalDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, RegistryFileName), nil
}

// Load reads the registry from disk. Returns an empty registry if the file
// does not exist.
func Load() (*Registry, error) {
	path, err := RegistryPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Registry{}, nil
		}
		return nil, fmt.Errorf("reading registry: %w", err)
	}

	var reg Registry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parsing registry: %w", err)
	}
	return &reg, nil
}

// Save writes the registry to disk, creating parent directories as needed.
func (r *Registry) Save() error {
	path, err := RegistryPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating registry directory: %w", err)
	}

	data, err := yaml.Marshal(r)
	if err != nil {
		return fmt.Errorf("marshalling registry: %w", err)
	}

	header := []byte("# Auto-managed by beans. Do not edit manually.\n")
	return os.WriteFile(path, append(header, data...), 0o644)
}

// Register adds a project to the registry and creates its local beans
// directory. If the project path is already registered, the existing entry
// is returned unchanged. The remoteURL is the git remote URL used to derive
// the slug; if empty, projectName or the basename of projectPath is used.
func (r *Registry) Register(projectPath, projectName, remoteURL string) (*ProjectEntry, error) {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, fmt.Errorf("resolving project path: %w", err)
	}

	// Check if already registered.
	if entry := r.Lookup(absPath); entry != nil {
		return entry, nil
	}

	slug := makeSlug(projectName, absPath, remoteURL, r)

	entry := ProjectEntry{
		Path:         absPath,
		Slug:         slug,
		RemoteURL:    remoteURL,
		RegisteredAt: time.Now().UTC().Truncate(time.Second),
	}
	r.Projects = append(r.Projects, entry)

	// Create the project beans directory.
	if err := r.ensureProjectDir(slug); err != nil {
		// Roll back the append.
		r.Projects = r.Projects[:len(r.Projects)-1]
		return nil, err
	}

	return &entry, nil
}

// Unregister removes a project from the registry by its absolute path.
// Returns true if the project was found and removed, false otherwise.
// Does NOT delete the project directory on disk.
func (r *Registry) Unregister(projectPath string) bool {
	absPath, _ := filepath.Abs(projectPath)
	for i, p := range r.Projects {
		if p.Path == absPath {
			r.Projects = append(r.Projects[:i], r.Projects[i+1:]...)
			return true
		}
	}
	return false
}

// Lookup finds a project entry by its absolute path. Returns nil if not found.
func (r *Registry) Lookup(projectPath string) *ProjectEntry {
	absPath, _ := filepath.Abs(projectPath)
	for i := range r.Projects {
		if r.Projects[i].Path == absPath {
			return &r.Projects[i]
		}
	}
	return nil
}

// ProjectDir returns the absolute path to a project's local beans root
// (the directory containing .beans.yml and .beans/).
func (r *Registry) ProjectDir(slug string) (string, error) {
	dir, err := LocalDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ProjectsDir, slug), nil
}

// ProjectBeansDir returns the absolute path to a project's .beans/ directory.
func (r *Registry) ProjectBeansDir(slug string) (string, error) {
	projDir, err := r.ProjectDir(slug)
	if err != nil {
		return "", err
	}
	return filepath.Join(projDir, ".beans"), nil
}

// ensureProjectDir creates the project directory structure under projects/.
func (r *Registry) ensureProjectDir(slug string) error {
	beansDir, err := r.ProjectBeansDir(slug)
	if err != nil {
		return err
	}
	return os.MkdirAll(beansDir, 0o755)
}

// makeSlug generates a unique slug for the project. If remoteURL is provided,
// derives the slug from the URL path (owner-repo). Otherwise falls back to
// projectName or the basename of projectPath. Appends a short hash suffix on
// collision.
func makeSlug(projectName, projectPath, remoteURL string, r *Registry) string {
	var base string

	if remoteURL != "" {
		base = slugFromRemoteURL(remoteURL)
	}

	if base == "" {
		name := projectName
		if name == "" {
			name = filepath.Base(projectPath)
		}
		base = bean.Slugify(name)
	}

	if base == "" {
		base = "project"
	}

	// Check for collision.
	if !slugExists(r, base) {
		return base
	}

	// Append short hash of the project path to make it unique.
	hash := sha256.Sum256([]byte(projectPath))
	suffix := fmt.Sprintf("%x", hash[:2]) // 4 hex chars
	return base + "-" + suffix
}

// sshRemoteRe matches SSH-style git remotes like git@github.com:owner/repo.git
var sshRemoteRe = regexp.MustCompile(`^[^@]+@([^:]+):(.+)$`)

// slugFromRemoteURL extracts an owner-repo slug from a git remote URL.
// Handles both HTTPS (https://github.com/owner/repo.git) and SSH
// (git@github.com:owner/repo.git) formats. Returns "" if parsing fails.
func slugFromRemoteURL(remoteURL string) string {
	var pathPart string

	if m := sshRemoteRe.FindStringSubmatch(remoteURL); m != nil {
		// SSH format: git@github.com:owner/repo.git
		pathPart = m[2]
	} else if u, err := url.Parse(remoteURL); err == nil && u.Host != "" {
		// HTTPS format: https://github.com/owner/repo.git
		pathPart = strings.TrimPrefix(u.Path, "/")
	}

	if pathPart == "" {
		return ""
	}

	// Strip .git suffix.
	pathPart = strings.TrimSuffix(pathPart, ".git")

	// Use the last two path segments (owner/repo) for the slug.
	parts := strings.Split(pathPart, "/")
	if len(parts) >= 2 {
		owner := parts[len(parts)-2]
		repo := parts[len(parts)-1]
		return bean.Slugify(owner + "-" + repo)
	}

	// Single segment (unlikely but handle gracefully).
	return bean.Slugify(parts[len(parts)-1])
}

func slugExists(r *Registry, slug string) bool {
	for _, p := range r.Projects {
		if p.Slug == slug {
			return true
		}
	}
	return false
}

func expandHome(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[1:]), nil
	}
	if filepath.IsAbs(path) {
		return path, nil
	}
	return filepath.Abs(path)
}
