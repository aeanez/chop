package filters

import (
	"strings"
	"testing"
)

var gitFetchFixture = `remote: Enumerating objects: 15, done.
remote: Counting objects: 100% (15/15), done.
remote: Compressing objects: 100% (8/8), done.
remote: Total 10 (delta 5), reused 7 (delta 2), pack-reused 0
Unpacking objects: 100% (10/10), 3.45 KiB | 352.00 KiB/s, done.
From https://github.com/user/repo
   abc1234..def5678  main       -> origin/main
 * [new branch]      feature/x  -> origin/feature/x
   111aaaa..222bbbb  develop    -> origin/develop`

var gitFetchNewTagFixture = `From https://github.com/user/repo
 * [new tag]         v1.5.0     -> v1.5.0`

var gitFetchPruneFixture = `From https://github.com/user/repo
 - [deleted]         (none)     -> origin/old-branch
   abc1234..def5678  main       -> origin/main`

var gitFetchUpToDateFixture = ``

func TestGitFetchStripsProgress(t *testing.T) {
	got, err := filterGitFetch(gitFetchFixture)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, "From https://github.com") {
		t.Errorf("expected remote URL in output: %s", got)
	}
	if !strings.Contains(got, "-> origin/main") {
		t.Errorf("expected ref update in output: %s", got)
	}
	if !strings.Contains(got, "[new branch]") {
		t.Errorf("expected new branch marker: %s", got)
	}
	if !strings.Contains(got, "-> origin/develop") {
		t.Errorf("expected develop ref update: %s", got)
	}

	// Noise stripped
	if strings.Contains(got, "Enumerating") {
		t.Errorf("Enumerating should be stripped: %s", got)
	}
	if strings.Contains(got, "Counting") {
		t.Errorf("Counting should be stripped: %s", got)
	}
	if strings.Contains(got, "Compressing") {
		t.Errorf("Compressing should be stripped: %s", got)
	}
	if strings.Contains(got, "Unpacking") {
		t.Errorf("Unpacking should be stripped: %s", got)
	}

	rawTokens := countTokens(gitFetchFixture)
	filteredTokens := countTokens(got)
	savings := 100.0 - (float64(filteredTokens)/float64(rawTokens))*100.0
	if savings < 40.0 {
		t.Errorf("expected >=40%% savings, got %.1f%%", savings)
	}
	t.Logf("token savings: %.1f%% (%d -> %d)", savings, rawTokens, filteredTokens)
	t.Logf("output:\n%s", got)
}

func TestGitFetchNewTag(t *testing.T) {
	got, err := filterGitFetch(gitFetchNewTagFixture)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, "[new tag]") {
		t.Errorf("expected new tag marker: %s", got)
	}
	if !strings.Contains(got, "v1.5.0") {
		t.Errorf("expected tag name: %s", got)
	}
	t.Logf("output:\n%s", got)
}

func TestGitFetchPrune(t *testing.T) {
	got, err := filterGitFetch(gitFetchPruneFixture)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, "[deleted]") {
		t.Errorf("expected deleted marker: %s", got)
	}
	if !strings.Contains(got, "-> origin/main") {
		t.Errorf("expected ref update: %s", got)
	}
	t.Logf("output:\n%s", got)
}

func TestGitFetchEmpty(t *testing.T) {
	got, err := filterGitFetch(gitFetchUpToDateFixture)
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty output for no-op fetch, got: %q", got)
	}
}

func TestGitFetchRouted(t *testing.T) {
	f := getGitFilter([]string{"fetch"})
	if f == nil {
		t.Fatal("expected filter for git fetch, got nil")
	}
}
