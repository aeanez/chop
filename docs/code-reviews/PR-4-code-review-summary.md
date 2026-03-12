# Code Review: PR-4

**Verdict:** ⚠️ CHANGES REQUESTED

| | |
| - | - |
| **Branch** | `feat/changelog` |
| **Title** | feat: add changelog generation and `chop changelog` command |
| **Author** | @aeanez (Andrés) |
| **Files Changed** | 5 |
| **Lines Changed** | +289 / -152 |
| **Date** | 2026-03-12 |

---

## Summary

This PR introduces `git-cliff` for automated changelog generation, a `cliff.toml` config, a `chop changelog` command (with `--latest`), and updates the release workflow to use `git-cliff-action`. The feature idea is solid — embedding the changelog in the binary is a genuinely nice touch. However, the approach conflicts with the hand-written CHANGELOG strategy that was just established on `main` today, and two HIGH-severity issues (undeclared dependency, floating action tag) need to be addressed before merging.

---

## Findings Overview

| Severity | In Scope | Out of Scope |
| -------- | -------- | ------------ |
| 🔴 CRITICAL | 0 | 0 |
| 🟠 HIGH | 3 | 0 |
| 🟡 MEDIUM | 3 | 0 |
| 🟢 LOW | 2 | 0 |
| ℹ️ INFO | 1 | 0 |

---

## In Scope Findings

### 🟠 HIGH-001: Architectural conflict with hand-written CHANGELOG on main

**Domains:** [Architecture]
**Location:** `.github/workflows/release.yml:44`

A hand-written CHANGELOG approach was just committed to `main` today (commit `dffd0c8`). That approach treats `CHANGELOG.md` as a human-authored document and fails the release build if no entry exists for the tag. This PR replaces it with auto-generation from `git-cliff`, which is a fundamentally different strategy.

Both can't coexist. The PR needs a conscious decision: commit to auto-generation (this PR) or stay with hand-written (current main). The auto-gen approach trades prose quality for zero-effort maintenance — the generated entries are noticeably weaker ("Add user-defined custom filters via filters.yml" repeated 3 times across separate commits vs. the detailed feature bullets in the original).

**Recommendation:**
Align on the strategy before merging. If auto-generation is the choice, the `--strip header` arg in the cliff action and the commit body template in `cliff.toml` should deduplicate entries per version (consider squashing related commits before tagging, or using `--tag-pattern` with cliff to group correctly).

---

### 🟠 HIGH-002: `git-cliff` undeclared build dependency

**Domains:** [Architecture, Code Quality]
**Location:** `Makefile:42`

The `changelog`, `release-patch`, `release-minor`, and `release-major` targets now silently require `git-cliff` to be installed on the host. If it isn't, the commands fail with a `command not found` error mid-way through the release process — after potentially writing partial state.

```makefile
# Current — fails silently if git-cliff not found
release-patch:
    @NEXT=...; \
    git-cliff --tag $$NEXT --output CHANGELOG.md && \
    git add CHANGELOG.md && git commit ...
```

**Recommendation:**
Add an install check at the top of the affected targets:

```makefile
release-patch:
    @command -v git-cliff >/dev/null 2>&1 || { echo "git-cliff is required: https://git-cliff.org/docs/installation"; exit 1; }
    ...
```

Also document the dependency in the README or a `CONTRIBUTING.md`.

---

### 🟠 HIGH-003: Floating action tag — supply chain risk

**Domains:** [Security]
**Location:** `.github/workflows/release.yml:35`

```yaml
uses: orhun/git-cliff-action@v4
```

Using a floating tag means any update to that action (including a compromised one) applies automatically to the release pipeline. The release job has `contents: write` permission and access to `GITHUB_TOKEN`, making it a sensitive target.

**Recommendation:**
Pin to a specific commit SHA:

```yaml
uses: orhun/git-cliff-action@14571d0b2f6b7e8b0e1c3f9d8c7a6b5e4d3c2f1a  # v4.x.x
```

---

### 🟡 MEDIUM-001: Release targets commit + tag + push in a single chain — no rollback

**Domains:** [Architecture]
**Location:** `Makefile:55-60`

```makefile
git-cliff --tag $$NEXT --output CHANGELOG.md && \
git add CHANGELOG.md && \
git commit -m "chore: update changelog for $$NEXT" && \
git tag $$NEXT && git push origin HEAD $$NEXT
```

If `git push` fails after the local tag is created, the repo is left in an inconsistent state (local commit + tag exist, remote has neither). Re-running the target would fail because the local tag already exists.

**Recommendation:**
Push before tagging, or add cleanup on failure:

```makefile
git tag $$NEXT || { echo "Tag already exists, aborting"; exit 1; } && \
git push origin HEAD $$NEXT || { git tag -d $$NEXT; exit 1; }
```

---

### 🟡 MEDIUM-002: `filter_unconventional = true` silently drops commits

**Domains:** [Code Quality]
**Location:** `cliff.toml:31`

```toml
filter_unconventional = true
```

Any commit that doesn't follow conventional commit format (`type: message`) is silently excluded from the changelog. This is a correctness risk — a meaningful fix committed as `"fix panic on nil pointer"` instead of `"fix: fix panic on nil pointer"` disappears with no warning.

**Recommendation:**
Either set `filter_unconventional = false` and add an `Other` catch-all group, or enforce conventional commits at the PR level (e.g., a CI check on PR titles). Silently dropping is worse than showing a messy entry.

---

### 🟡 MEDIUM-003: Embedded CHANGELOG can be stale on local builds

**Domains:** [Code Quality]
**Location:** `main.go:24`

```go
//go:embed CHANGELOG.md
var changelog string
```

`make build` doesn't depend on `make changelog`, so a developer running `make build` locally will embed whatever CHANGELOG.md was last generated. The `chop changelog` output will be stale until they manually run `make changelog` first.

**Recommendation:**
Add a `build` dependency:

```makefile
build: changelog
    docker compose run --rm ...
```

Or add a comment warning developers to regenerate before building for release.

---

### 🟢 LOW-001: `chop changelog --latest` shows `[Unreleased]` section

**Domains:** [Code Quality]
**Location:** `main.go:988`

`extractLatestVersion` finds the first `## [` section, which will be `## [Unreleased]` when there are commits since the last tag. A user running `chop changelog --latest` after making local commits would see unreleased changes rather than the installed version's changes.

**Recommendation:**
Skip the `[Unreleased]` section, or label the output clearly ("Changes in this build: ...").

---

### 🟢 LOW-002: CHANGELOG version header format inconsistency

**Domains:** [Code Quality]
**Location:** `cliff.toml:16`

`cliff.toml` strips the `v` prefix: `{{ version | trim_start_matches(pat="v") }}` → `## [1.6.0]`. The old hand-written CHANGELOG used `## [v1.6.0]`. The generated file now mixes both formats in history (pre-v1.0.0 entries kept from the old format, new entries without `v`).

**Recommendation:**
Decide on one format. The conventional convention for CHANGELOG.md is `[1.6.0]` without `v` (Keep a Changelog spec), so the cliff config is correct — but the old entries should be cleaned up for consistency.

---

### ℹ️ INFO-001: `chop changelog` embedded binary is a genuinely useful addition

**Domains:** [Architecture]
**Location:** `main.go:24`

Embedding the CHANGELOG in the binary so users can run `chop changelog --latest` to see what changed in their installed version is a good UX pattern. Worth keeping regardless of which changelog generation strategy is chosen.

---

## Action Items

### Must Fix (blocks merge)

- [ ] **HIGH-001** — Resolve the architectural conflict: align on auto-generated (this PR) vs. hand-written (current main) before merging
- [ ] **HIGH-002** — Add `git-cliff` install check to Makefile release targets and document the dependency
- [ ] **HIGH-003** — Pin `orhun/git-cliff-action` to a commit SHA

### Should Fix

- [ ] **MEDIUM-001** — Guard the release tag + push chain against partial-failure state
- [ ] **MEDIUM-002** — Handle unconventional commits explicitly rather than silently dropping them
- [ ] **MEDIUM-003** — Make `build` depend on `changelog`, or document the manual step

### Consider

- [ ] **LOW-001** — Skip `[Unreleased]` in `--latest` output
- [ ] **LOW-002** — Normalize CHANGELOG version header format across history

---

## Files Reviewed

| File | Findings |
| ---- | -------- |
| `.github/workflows/release.yml` | HIGH-001, HIGH-003 |
| `Makefile` | HIGH-001, HIGH-002, MEDIUM-001 |
| `cliff.toml` | MEDIUM-002 |
| `main.go` | MEDIUM-003, LOW-001, INFO-001 |
| `CHANGELOG.md` | LOW-002 |

---

🤖 Generated with [Claude Code](https://claude.com/claude-code)
