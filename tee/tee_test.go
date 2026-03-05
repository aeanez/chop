package tee

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("CHOP_TEE_DIR", dir)
	t.Setenv("CHOP_TEE", "") // ensure not disabled
	return dir
}

func longString(n int) string {
	return strings.Repeat("x", n)
}

func TestSaveOnFailure(t *testing.T) {
	dir := setupTestDir(t)

	raw := longString(600)
	path := Save("git status", raw, 1, 10.0) // exitCode=1 → failure

	if path == "" {
		t.Fatal("expected file to be created on failure, got empty path")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}
	if string(content) != raw {
		t.Error("saved content does not match raw input")
	}

	// Verify file is in the tee dir
	if !strings.HasPrefix(path, dir) {
		t.Errorf("expected path under %s, got %s", dir, path)
	}
}

func TestSaveOnHighSavings(t *testing.T) {
	setupTestDir(t)

	raw := longString(600)
	path := Save("cargo test", raw, 0, 85.0) // success but >80% savings

	if path == "" {
		t.Fatal("expected file to be created on high savings, got empty path")
	}
}

func TestSkipOnLowSavingsAndSuccess(t *testing.T) {
	setupTestDir(t)

	raw := longString(600)
	path := Save("git log", raw, 0, 50.0) // success, low savings

	if path != "" {
		t.Errorf("expected empty path for success+low savings, got %s", path)
	}
}

func TestSkipOnShortOutput(t *testing.T) {
	setupTestDir(t)

	raw := longString(100) // < 500 chars
	path := Save("git status", raw, 1, 90.0)

	if path != "" {
		t.Errorf("expected empty path for short output, got %s", path)
	}
}

func TestSkipWhenDisabled(t *testing.T) {
	setupTestDir(t)
	t.Setenv("CHOP_TEE", "off")

	raw := longString(600)
	path := Save("git status", raw, 1, 90.0)

	if path != "" {
		t.Errorf("expected empty path when CHOP_TEE=off, got %s", path)
	}
}

func TestRotation(t *testing.T) {
	dir := setupTestDir(t)

	// Create 25 files
	for i := 0; i < 25; i++ {
		name := filepath.Join(dir, strings.Repeat("a", 3)+"-"+string(rune('a'+i))+".txt")
		if err := os.WriteFile(name, []byte("data"), 0o644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	// Save should trigger rotation
	raw := longString(600)
	path := Save("test cmd", raw, 1, 90.0)

	if path == "" {
		t.Fatal("expected file to be created after rotation")
	}

	// Count remaining .txt files
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}
	count := 0
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".txt") {
			count++
		}
	}

	if count > 20 {
		t.Errorf("expected at most 20 files after rotation, got %d", count)
	}
}

func TestTruncation(t *testing.T) {
	dir := setupTestDir(t)

	// Create content larger than 1MB
	raw := longString(maxBytes + 5000)
	path := Save("big cmd", raw, 1, 90.0)

	if path == "" {
		t.Fatal("expected file to be created for large content")
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}
	if info.Size() > int64(maxBytes) {
		t.Errorf("expected file size <= %d, got %d", maxBytes, info.Size())
	}

	// Verify file is in dir
	if !strings.HasPrefix(path, dir) {
		t.Errorf("expected path under %s, got %s", dir, path)
	}
}

func TestSanitize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"git status", "git-status"},
		{"git log --oneline -10", "git-log-oneline-10"},
		{"", "cmd"},
	}
	for _, tt := range tests {
		got := sanitize(tt.input)
		if got != tt.want {
			t.Errorf("sanitize(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
