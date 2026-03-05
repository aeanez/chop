package filters

import (
	"strings"
	"testing"
)

func TestNpmTestAllPassing(t *testing.T) {
	raw := ` PASS  src/utils.test.ts
 PASS  src/auth/login.test.ts
 PASS  src/api/users.test.ts
 PASS  src/api/posts.test.ts
 PASS  src/components/Header.test.tsx
 PASS  src/components/Footer.test.tsx
 PASS  src/hooks/useAuth.test.ts
 PASS  src/hooks/useQuery.test.ts

Test Suites: 8 passed, 8 total
Tests:       42 passed, 42 total
Snapshots:   0 total
Time:        4.231 s
Ran all test suites.
`
	got, err := filterNpmTestCmd(raw)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, "all 42 tests passed") {
		t.Errorf("expected 'all 42 tests passed', got: %s", got)
	}

	rawTokens := countTokens(raw)
	filteredTokens := countTokens(got)
	savings := 100.0 - (float64(filteredTokens) / float64(rawTokens) * 100.0)
	if savings < 70.0 {
		t.Errorf("expected >=70%% savings, got %.1f%%", savings)
	}
	t.Logf("token savings: %.1f%% (%d -> %d)", savings, rawTokens, filteredTokens)
	t.Logf("output: %s", got)
}

func TestNpmTestWithFailures(t *testing.T) {
	raw := ` PASS  src/utils.test.ts
 PASS  src/auth/login.test.ts
 FAIL  src/api/users.test.ts
  ● getUser > should return user by id

    expect(received).toEqual(expected)

    Expected: {"id": 1, "name": "John"}
    Received: {"id": 1, "name": "Jane"}

      12 |   const user = await getUser(1);
      13 |   expect(user).toEqual({ id: 1, name: "John" });
         |                ^
      14 | });

 PASS  src/api/posts.test.ts
 FAIL  src/components/Header.test.tsx
  ● Header > should render title

    Error: Unable to find element with text: Welcome

      at Object.<anonymous> (src/components/Header.test.tsx:8:5)

 PASS  src/hooks/useAuth.test.ts
 PASS  src/hooks/useQuery.test.ts

Test Suites: 2 failed, 5 passed, 7 total
Tests:       2 failed, 35 passed, 37 total
Snapshots:   0 total
Time:        6.882 s
Ran all test suites.
`
	got, err := filterNpmTestCmd(raw)
	if err != nil {
		t.Fatal(err)
	}

	// Failures should be shown
	if !strings.Contains(got, "FAIL") {
		t.Errorf("expected FAIL markers, got: %s", got)
	}
	if !strings.Contains(got, "should return user by id") || !strings.Contains(got, "should render title") {
		t.Errorf("expected failure details, got: %s", got)
	}

	// Summary should be present
	if !strings.Contains(got, "2 failed") {
		t.Errorf("expected failure count in summary, got: %s", got)
	}
	if !strings.Contains(got, "35 passed") {
		t.Errorf("expected pass count in summary, got: %s", got)
	}

	// PASS lines should be stripped
	if strings.Contains(got, "PASS  src/utils.test.ts") {
		t.Errorf("expected PASS lines stripped, got: %s", got)
	}

	rawTokens := countTokens(raw)
	filteredTokens := countTokens(got)
	savings := 100.0 - (float64(filteredTokens) / float64(rawTokens) * 100.0)
	if savings < 30.0 {
		t.Errorf("expected >=30%% savings, got %.1f%%", savings)
	}
	t.Logf("token savings: %.1f%% (%d -> %d)", savings, rawTokens, filteredTokens)
	t.Logf("output:\n%s", got)
}

func TestNpmTestMochaStyle(t *testing.T) {
	raw := `
  User API
    ✓ should create user (45ms)
    ✓ should get user by id (12ms)
    ✓ should update user (23ms)
    ✓ should delete user (18ms)

  Auth API
    ✓ should login (67ms)
    ✓ should logout (8ms)
    ✓ should refresh token (34ms)

  7 passing (207ms)
`
	got, err := filterNpmTestCmd(raw)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, "all 7 tests passed") {
		t.Errorf("expected 'all 7 tests passed', got: %s", got)
	}

	rawTokens := countTokens(raw)
	filteredTokens := countTokens(got)
	savings := 100.0 - (float64(filteredTokens) / float64(rawTokens) * 100.0)
	if savings < 70.0 {
		t.Errorf("expected >=70%% savings, got %.1f%%", savings)
	}
	t.Logf("token savings: %.1f%% (%d -> %d)", savings, rawTokens, filteredTokens)
}
