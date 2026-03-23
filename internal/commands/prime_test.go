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
	entry, err := reg.Register(projectDir, "test-project")
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
