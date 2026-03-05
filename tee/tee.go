package tee

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	maxFiles   = 20
	maxBytes   = 1024 * 1024 // 1MB
	minChars   = 500
	teeDirName = "tee"
)

var sanitizeRe = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

// Dir returns the tee directory path.
func Dir() string {
	if d := os.Getenv("CHOP_TEE_DIR"); d != "" {
		return d
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".local", "share", "chop", teeDirName)
}

// Save saves raw output to a file for later retrieval by LLMs.
// Returns the file path on success, or "" silently on any error or skip condition.
// Tee must never break the tool — all errors are swallowed.
func Save(command string, raw string, exitCode int, savingsPct float64) string {
	// Check if tee is disabled
	if strings.EqualFold(os.Getenv("CHOP_TEE"), "off") {
		return ""
	}

	// Skip short output
	if len(raw) < minChars {
		return ""
	}

	// Only save on failure or high savings
	if exitCode == 0 && savingsPct <= 80.0 {
		return ""
	}

	dir := Dir()

	// Create directory if needed
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return ""
	}

	// Rotate old files
	rotate(dir)

	// Build filename
	name := sanitize(command)
	ts := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s-%s.txt", name, ts)
	path := filepath.Join(dir, filename)

	// Truncate content if too large
	content := raw
	if len(content) > maxBytes {
		content = content[:maxBytes]
	}

	// Write file
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return ""
	}

	return path
}

func sanitize(command string) string {
	s := sanitizeRe.ReplaceAllString(command, "-")
	// Collapse multiple hyphens
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")
	if len(s) > 60 {
		s = s[:60]
	}
	if s == "" {
		s = "cmd"
	}
	return s
}

func rotate(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	// Filter to only .txt files
	var files []os.DirEntry
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".txt") {
			files = append(files, e)
		}
	}

	// If under limit, nothing to do (we need room for 1 new file)
	if len(files) < maxFiles {
		return
	}

	// Sort by mod time ascending (oldest first)
	sort.Slice(files, func(i, j int) bool {
		fi, _ := files[i].Info()
		fj, _ := files[j].Info()
		if fi == nil || fj == nil {
			return files[i].Name() < files[j].Name()
		}
		return fi.ModTime().Before(fj.ModTime())
	})

	// Delete oldest files to make room (keep maxFiles-1 so new file fits)
	toDelete := len(files) - maxFiles + 1
	for i := 0; i < toDelete; i++ {
		_ = os.Remove(filepath.Join(dir, files[i].Name()))
	}
}
