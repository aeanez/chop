package filters

import (
	"strings"
	"testing"
)

func countTokens(s string) int {
	return len(strings.Fields(s))
}

func TestGitStatusClean(t *testing.T) {
	raw := `On branch main
Your branch is up to date with 'origin/main'.

nothing to commit, working tree clean
`
	got, err := filterGitStatus(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got != "clean" {
		t.Errorf("expected 'clean', got %q", got)
	}
}

func TestGitStatusModified(t *testing.T) {
	raw := `On branch feature/login
Your branch is up to date with 'origin/feature/login'.

Changes not staged for commit:
  (use "git add <file>..." to update what will be committed)
  (use "git restore <file>..." to discard changes in working directory)
	modified:   src/app.ts
	modified:   src/auth/login.ts
	deleted:    src/old-config.json

Untracked files:
  (use "git add <file>..." to include in what will be committed)
	src/new-feature.ts
	docs/notes.md

no changes added to commit (use "git add" to track)
`
	got, err := filterGitStatus(raw)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, "modified(3)") {
		t.Errorf("expected modified count 3, got: %s", got)
	}
	if !strings.Contains(got, "untracked(2)") {
		t.Errorf("expected untracked count 2, got: %s", got)
	}

	// Verify token savings
	rawTokens := countTokens(raw)
	filteredTokens := countTokens(got)
	savings := 100.0 - (float64(filteredTokens) / float64(rawTokens) * 100.0)
	if savings < 60.0 {
		t.Errorf("expected >=60%% savings, got %.1f%%", savings)
	}
	t.Logf("token savings: %.1f%% (%d -> %d)", savings, rawTokens, filteredTokens)
}

func TestGitStatusStaged(t *testing.T) {
	raw := `On branch main
Changes to be committed:
  (use "git restore --staged <file>..." to unstage)
	new file:   src/feature.go
	new file:   src/feature_test.go

Changes not staged for commit:
  (use "git add <file>..." to update what will be committed)
	modified:   README.md
`
	got, err := filterGitStatus(raw)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, "staged(2)") {
		t.Errorf("expected staged count 2, got: %s", got)
	}
	if !strings.Contains(got, "modified(1)") {
		t.Errorf("expected modified count 1, got: %s", got)
	}
}
