package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFrom_MissingFile(t *testing.T) {
	cfg := LoadFrom("/nonexistent/path/config.yml")
	if cfg.Tee != false {
		t.Error("expected Tee=false for missing file")
	}
	if len(cfg.Disabled) != 0 {
		t.Error("expected empty Disabled for missing file")
	}
}

func TestLoadFrom_EmptyFile(t *testing.T) {
	tmp := writeTemp(t, "")
	cfg := LoadFrom(tmp)
	if cfg.Tee != false {
		t.Error("expected Tee=false for empty file")
	}
	if len(cfg.Disabled) != 0 {
		t.Error("expected empty Disabled for empty file")
	}
}

func TestLoadFrom_CommentsOnly(t *testing.T) {
	content := "# this is a comment\n# another comment\n"
	tmp := writeTemp(t, content)
	cfg := LoadFrom(tmp)
	if cfg.Tee != false {
		t.Error("expected Tee=false for comments-only file")
	}
}

func TestLoadFrom_TeeTrue(t *testing.T) {
	content := "tee: true\n"
	tmp := writeTemp(t, content)
	cfg := LoadFrom(tmp)
	if cfg.Tee != true {
		t.Error("expected Tee=true")
	}
}

func TestLoadFrom_TeeFalse(t *testing.T) {
	content := "tee: false\n"
	tmp := writeTemp(t, content)
	cfg := LoadFrom(tmp)
	if cfg.Tee != false {
		t.Error("expected Tee=false")
	}
}

func TestLoadFrom_DisabledList(t *testing.T) {
	content := "disabled: [git, docker]\n"
	tmp := writeTemp(t, content)
	cfg := LoadFrom(tmp)
	if len(cfg.Disabled) != 2 {
		t.Fatalf("expected 2 disabled items, got %d", len(cfg.Disabled))
	}
	if cfg.Disabled[0] != "git" {
		t.Errorf("expected 'git', got %q", cfg.Disabled[0])
	}
	if cfg.Disabled[1] != "docker" {
		t.Errorf("expected 'docker', got %q", cfg.Disabled[1])
	}
}

func TestLoadFrom_DisabledEmpty(t *testing.T) {
	content := "disabled: []\n"
	tmp := writeTemp(t, content)
	cfg := LoadFrom(tmp)
	if len(cfg.Disabled) != 0 {
		t.Errorf("expected empty disabled list, got %d", len(cfg.Disabled))
	}
}

func TestLoadFrom_DisabledWithQuotes(t *testing.T) {
	content := `disabled: ["git", "docker"]` + "\n"
	tmp := writeTemp(t, content)
	cfg := LoadFrom(tmp)
	if len(cfg.Disabled) != 2 {
		t.Fatalf("expected 2 disabled items, got %d", len(cfg.Disabled))
	}
	if cfg.Disabled[0] != "git" {
		t.Errorf("expected 'git', got %q", cfg.Disabled[0])
	}
}

func TestLoadFrom_FullConfig(t *testing.T) {
	content := `# chop config
tee: true
disabled: [git, docker, kubectl]
`
	tmp := writeTemp(t, content)
	cfg := LoadFrom(tmp)
	if cfg.Tee != true {
		t.Error("expected Tee=true")
	}
	if len(cfg.Disabled) != 3 {
		t.Fatalf("expected 3 disabled items, got %d", len(cfg.Disabled))
	}
}

func TestLoadFrom_InlineComments(t *testing.T) {
	content := "tee: true  # enable tee\ndisabled: [git] # skip git\n"
	tmp := writeTemp(t, content)
	cfg := LoadFrom(tmp)
	if cfg.Tee != true {
		t.Error("expected Tee=true with inline comment")
	}
	if len(cfg.Disabled) != 1 || cfg.Disabled[0] != "git" {
		t.Errorf("expected [git], got %v", cfg.Disabled)
	}
}

func TestIsDisabled(t *testing.T) {
	cfg := Config{Disabled: []string{"git", "docker"}}

	if !cfg.IsDisabled("git") {
		t.Error("expected git to be disabled")
	}
	if !cfg.IsDisabled("Git") {
		t.Error("expected Git (case-insensitive) to be disabled")
	}
	if !cfg.IsDisabled("docker") {
		t.Error("expected docker to be disabled")
	}
	if cfg.IsDisabled("npm") {
		t.Error("expected npm to NOT be disabled")
	}
}

func TestIsDisabled_Empty(t *testing.T) {
	cfg := Config{}
	if cfg.IsDisabled("git") {
		t.Error("expected nothing disabled on empty config")
	}
}

func TestPath_Default(t *testing.T) {
	// Unset XDG to test default
	old := os.Getenv("XDG_CONFIG_HOME")
	os.Unsetenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", old)

	p := Path()
	if !filepath.IsAbs(p) {
		// On most systems this will be absolute via UserHomeDir
		// Just check it ends correctly
	}
	if filepath.Base(p) != "config.yml" {
		t.Errorf("expected config.yml, got %s", filepath.Base(p))
	}
}

func TestPath_XDG(t *testing.T) {
	old := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", "/tmp/xdg")
	defer os.Setenv("XDG_CONFIG_HOME", old)

	p := Path()
	expected := filepath.Join("/tmp/xdg", "chop", "config.yml")
	if p != expected {
		t.Errorf("expected %s, got %s", expected, p)
	}
}

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
