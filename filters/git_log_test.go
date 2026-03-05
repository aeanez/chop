package filters

import (
	"strings"
	"testing"
)

func TestGitLogVerbose(t *testing.T) {
	raw := `commit a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2
Author: Alice Johnson <alice.johnson@bigcorp-engineering.com>
Date:   Mon Mar 3 10:00:00 2026 -0500

    Add user authentication module

    Implements JWT-based auth with refresh tokens.
    Includes middleware for route protection.

commit b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3
Author: Bob Williams <bob.williams@bigcorp-engineering.com>
Date:   Sun Mar 2 15:30:00 2026 -0500

    Fix database connection pooling issue

    The pool was exhausted under high load because
    connections were not properly returned after timeout.

commit c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4
Author: Charlie Brown <charlie.brown@bigcorp-engineering.com>
Date:   Sat Mar 1 09:15:00 2026 -0500

    Refactor payment processing pipeline

    Split monolithic handler into smaller services.
    Added retry logic for failed transactions.

commit d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5
Author: Diana Martinez <diana.martinez@bigcorp-engineering.com>
Date:   Fri Feb 28 18:45:00 2026 -0500

    Update dependencies to latest versions

    Bumped express 4.18->4.21, bcrypt 5.1->5.2.
    Ran full test suite, all passing.

commit e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6
Author: Eve Anderson <eve.anderson@bigcorp-engineering.com>
Date:   Thu Feb 27 12:00:00 2026 -0500

    Add integration tests for API endpoints

    Covers auth, payments, and user profile routes.
    Uses supertest with in-memory database.

commit f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1
Author: Frank Thompson <frank.thompson@bigcorp-engineering.com>
Date:   Wed Feb 26 08:30:00 2026 -0500

    Initial project setup

    Scaffolded Express app with TypeScript config.
    Added ESLint, Prettier, and Husky pre-commit hooks.
`

	got, err := filterGitLog(raw)
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) != 6 {
		t.Errorf("expected 6 condensed lines, got %d: %s", len(lines), got)
	}

	// Each line should be <short-hash> <message>
	for _, line := range lines {
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			t.Errorf("expected 'hash message' format, got: %s", line)
		}
		if len(parts[0]) != 7 {
			t.Errorf("expected 7-char short hash, got %d chars: %s", len(parts[0]), parts[0])
		}
	}

	// Verify first commit
	if !strings.Contains(lines[0], "a1b2c3d") {
		t.Errorf("first line should contain a1b2c3d, got: %s", lines[0])
	}
	if !strings.Contains(lines[0], "Add user authentication module") {
		t.Errorf("first line should contain commit message, got: %s", lines[0])
	}

	// Token savings >= 70%
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - (float64(filteredTokens)/float64(rawTokens))*100.0
	if savings < 70.0 {
		t.Errorf("expected >=70%% savings, got %.1f%%", savings)
	}
	t.Logf("token savings: %.1f%% (%d -> %d)", savings, rawTokens, filteredTokens)
}

func TestGitLogOnelinePassthrough(t *testing.T) {
	raw := `a1b2c3d Add user auth
b2c3d4e Fix db pooling
c3d4e5f Refactor payments`

	got, err := filterGitLog(raw)
	if err != nil {
		t.Fatal(err)
	}

	if got != raw {
		t.Errorf("oneline format should pass through unchanged, got: %s", got)
	}
}

func TestGitLogTruncation(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 25; i++ {
		hash := strings.Repeat("a", 40)
		b.WriteString("commit " + hash + "\n")
		b.WriteString("Author: Test <test@test.com>\n")
		b.WriteString("Date:   Mon Mar 3 10:00:00 2026 -0500\n")
		b.WriteString("\n")
		b.WriteString("    Commit number " + strings.Repeat("x", 5) + "\n\n")
	}

	got, err := filterGitLog(b.String())
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(got), "\n")
	// 20 entries + 1 "(5 more)" line
	if len(lines) != 21 {
		t.Errorf("expected 21 lines (20 + truncation notice), got %d", len(lines))
	}
	last := lines[len(lines)-1]
	if !strings.Contains(last, "(5 more)") {
		t.Errorf("expected truncation notice, got: %s", last)
	}
}

func TestGitLogEmpty(t *testing.T) {
	got, err := filterGitLog("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty output for empty input, got: %s", got)
	}
}
