package filters

import (
	"strings"
	"testing"
)

func TestFilterGitStatManyFiles(t *testing.T) {
	raw := ` src/api/handler.go         |  45 +++++++++++++++--
 src/api/middleware.go      |  12 ++--
 src/models/user.go         |  83 ++++++++++++++++++++++++++++++-----
 src/models/product.go      |   3 +
 src/config/settings.go     |  21 ++++-----
 tests/api_test.go          |  67 +++++++++++++++++++++++++----
 tests/models_test.go       |  34 +++++++++------
 docs/README.md             |   8 +-
 src/db/migrations.go       |  55 ++++++++++++++++++++---
 src/db/queries.go          |  19 +++++--
 src/handlers/auth.go       |  41 ++++++++++++------
 100 files changed, 1243 insertions(+), 387 deletions(-)`

	got, err := filterGitStat(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have summary line
	if !strings.Contains(got, "100 files changed") {
		t.Errorf("expected summary line, got:\n%s", got)
	}

	// Should have top files
	if !strings.Contains(got, "top:") {
		t.Errorf("expected 'top:' section, got:\n%s", got)
	}

	// Top file should be the one with most changes (user.go at 83)
	if !strings.Contains(got, "src/models/user.go") {
		t.Errorf("expected top file src/models/user.go, got:\n%s", got)
	}

	// Should not expand
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	if filteredTokens > rawTokens {
		t.Errorf("filter expanded output: raw=%d tokens, filtered=%d tokens", rawTokens, filteredTokens)
	}

	t.Logf("before:\n%s", raw)
	t.Logf("after:\n%s", got)
	t.Logf("tokens: %d -> %d", rawTokens, filteredTokens)
}

func TestFilterGitStatFewFilesPassthrough(t *testing.T) {
	raw := ` src/main.go | 12 +++---
 src/util.go |  4 ++
 2 files changed, 10 insertions(+), 6 deletions(-)`

	got, err := filterGitStat(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != raw {
		t.Errorf("expected passthrough for <=5 files, got:\n%s", got)
	}
}

func TestFilterGitStatViaFilterGitDiff(t *testing.T) {
	// Verify the integration: filterGitDiff delegates to filterGitStat
	raw := ` src/api/handler.go    |  45 ++++++++++++
 src/models/user.go    |  83 +++++++++++++++++++
 src/config/app.go     |  21 +++++
 tests/api_test.go     |  67 +++++++++++++++
 tests/unit_test.go    |  34 ++++++++
 src/db/migrations.go  |  55 ++++++++++++
 src/db/queries.go     |  19 ++++
 10 files changed, 324 insertions(+), 89 deletions(-)`

	got, err := filterGitDiff(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "10 files changed") {
		t.Errorf("expected summary in output, got:\n%s", got)
	}
	if !strings.Contains(got, "top:") {
		t.Errorf("expected top files section, got:\n%s", got)
	}
}
