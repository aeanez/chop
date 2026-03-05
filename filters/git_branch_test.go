package filters

import (
	"strings"
	"testing"
)

func TestGitBranchListing(t *testing.T) {
	raw := `  develop
  feature/auth
  feature/payments
  fix/login-bug
* main
  release/v1.0
  staging`

	got, err := filterGitBranch(raw)
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(got), "\n")

	// Current branch should be first
	if !strings.HasPrefix(lines[0], "* main") {
		t.Errorf("expected current branch first with *, got: %s", lines[0])
	}

	// Should have all 7 branches
	if len(lines) != 7 {
		t.Errorf("expected 7 branches, got %d: %s", len(lines), got)
	}

	// Other branches should not have *
	for _, line := range lines[1:] {
		if strings.HasPrefix(strings.TrimSpace(line), "*") {
			t.Errorf("only current branch should have *, got: %s", line)
		}
	}

	// Output should be compact
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			t.Error("unexpected empty line in output")
		}
	}
}

func TestGitBranchEmpty(t *testing.T) {
	got, err := filterGitBranch("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty output for empty input, got: %s", got)
	}
}

func TestGitBranchSingle(t *testing.T) {
	raw := `* main`

	got, err := filterGitBranch(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got != "* main" {
		t.Errorf("expected '* main', got: %s", got)
	}
}
