package filters

import (
	"strings"
	"testing"
)

var gitPullFastForwardFixture = `remote: Enumerating objects: 5, done.
remote: Counting objects: 100% (5/5), done.
remote: Compressing objects: 100% (3/3), done.
remote: Total 3 (delta 2), reused 0 (delta 0), pack-reused 0
Unpacking objects: 100% (3/3), 1.23 KiB | 615.00 KiB/s, done.
From https://github.com/user/repo
   abc1234..def5678  main       -> origin/main
Updating abc1234..def5678
Fast-forward
 src/app.ts    | 15 +++++++--------
 src/utils.ts  |  3 ++-
 2 files changed, 9 insertions(+), 9 deletions(-)`

var gitPullMergeFixture = `remote: Enumerating objects: 10, done.
remote: Counting objects: 100% (10/10), done.
remote: Compressing objects: 100% (6/6), done.
remote: Total 6 (delta 3), reused 0 (delta 0), pack-reused 0
Unpacking objects: 100% (6/6), 2.50 KiB | 425.00 KiB/s, done.
From https://github.com/user/repo
   abc1234..def5678  main       -> origin/main
Merge made by the 'ort' strategy.
 src/app.ts    | 15 +++++++--------
 src/utils.ts  |  3 ++-
 README.md     | 10 ++++------
 3 files changed, 12 insertions(+), 16 deletions(-)`

var gitPullRebaseFixture = `remote: Enumerating objects: 5, done.
remote: Counting objects: 100% (5/5), done.
remote: Compressing objects: 100% (3/3), done.
remote: Total 3 (delta 2), reused 0 (delta 0), pack-reused 0
Unpacking objects: 100% (3/3), done.
From https://github.com/user/repo
   abc1234..def5678  main       -> origin/main
Successfully rebased and updated refs/heads/main.`

var gitPullConflictFixture = `remote: Enumerating objects: 5, done.
remote: Counting objects: 100% (5/5), done.
remote: Total 3 (delta 2), reused 0 (delta 0)
Unpacking objects: 100% (3/3), done.
From https://github.com/user/repo
   abc1234..def5678  main       -> origin/main
Updating abc1234..def5678
Auto-merging src/app.ts
CONFLICT (content): Merge conflict in src/app.ts
Automatic merge failed; fix conflicts and then commit the result.`

var gitPullNewFilesFixture = `remote: Enumerating objects: 8, done.
remote: Counting objects: 100% (8/8), done.
remote: Compressing objects: 100% (4/4), done.
remote: Total 4 (delta 2), reused 0 (delta 0), pack-reused 0
Unpacking objects: 100% (4/4), done.
From https://github.com/user/repo
   abc1234..def5678  main       -> origin/main
Updating abc1234..def5678
Fast-forward
 src/new.ts | 20 ++++++++++++++++++++
 2 files changed, 20 insertions(+)
 create mode 100644 src/new.ts`

func TestGitPullAlreadyUpToDate(t *testing.T) {
	raw := "Already up to date."
	got, err := filterGitPull(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got != raw {
		t.Errorf("expected pass-through, got: %s", got)
	}
}

func TestGitPullFastForward(t *testing.T) {
	got, err := filterGitPull(gitPullFastForwardFixture)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, "From https://github.com") {
		t.Errorf("expected remote URL: %s", got)
	}
	if !strings.Contains(got, "Fast-forward") {
		t.Errorf("expected Fast-forward: %s", got)
	}
	if !strings.Contains(got, "2 files changed") {
		t.Errorf("expected diffstat summary: %s", got)
	}

	// Noise stripped
	if strings.Contains(got, "Enumerating") {
		t.Errorf("Enumerating should be stripped: %s", got)
	}
	if strings.Contains(got, "Unpacking") {
		t.Errorf("Unpacking should be stripped: %s", got)
	}
	if strings.Contains(got, "Updating abc") {
		t.Errorf("Updating hash line should be stripped: %s", got)
	}
	// Per-file diffstat stripped
	if strings.Contains(got, "src/app.ts") {
		t.Errorf("per-file diffstat should be stripped: %s", got)
	}

	rawTokens := countTokens(gitPullFastForwardFixture)
	filteredTokens := countTokens(got)
	savings := 100.0 - (float64(filteredTokens)/float64(rawTokens))*100.0
	if savings < 50.0 {
		t.Errorf("expected >=50%% savings, got %.1f%%", savings)
	}
	t.Logf("token savings: %.1f%% (%d -> %d)", savings, rawTokens, filteredTokens)
	t.Logf("output:\n%s", got)
}

func TestGitPullMerge(t *testing.T) {
	got, err := filterGitPull(gitPullMergeFixture)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, "Merge made by") {
		t.Errorf("expected merge strategy line: %s", got)
	}
	if !strings.Contains(got, "3 files changed") {
		t.Errorf("expected diffstat summary: %s", got)
	}
	if strings.Contains(got, "src/app.ts") {
		t.Errorf("per-file diffstat should be stripped: %s", got)
	}
	t.Logf("output:\n%s", got)
}

func TestGitPullRebase(t *testing.T) {
	got, err := filterGitPull(gitPullRebaseFixture)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, "Successfully rebased") {
		t.Errorf("expected rebase result: %s", got)
	}
	if strings.Contains(got, "Enumerating") {
		t.Errorf("progress should be stripped: %s", got)
	}
	t.Logf("output:\n%s", got)
}

func TestGitPullConflict(t *testing.T) {
	got, err := filterGitPull(gitPullConflictFixture)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, "CONFLICT") {
		t.Errorf("expected CONFLICT line: %s", got)
	}
	if !strings.Contains(got, "Auto-merging src/app.ts") {
		t.Errorf("expected Auto-merging line: %s", got)
	}
	if !strings.Contains(got, "Automatic merge failed") {
		t.Errorf("expected merge failed line: %s", got)
	}
	t.Logf("output:\n%s", got)
}

func TestGitPullNewFiles(t *testing.T) {
	got, err := filterGitPull(gitPullNewFilesFixture)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, "create mode") {
		t.Errorf("expected create mode line: %s", got)
	}
	if !strings.Contains(got, "2 files changed") {
		t.Errorf("expected diffstat summary: %s", got)
	}
	t.Logf("output:\n%s", got)
}

func TestGitPullEmpty(t *testing.T) {
	got, err := filterGitPull("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty output, got: %q", got)
	}
}

func TestGitPullRouted(t *testing.T) {
	f := getGitFilter([]string{"pull"})
	if f == nil {
		t.Fatal("expected filter for git pull, got nil")
	}
}

func TestGitFetchRoutedFromPull(t *testing.T) {
	f := getGitFilter([]string{"fetch"})
	if f == nil {
		t.Fatal("expected filter for git fetch, got nil")
	}
}