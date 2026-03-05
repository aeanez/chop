package filters

import (
	"strings"
	"testing"
)

func TestFilterGoTestAllPassing(t *testing.T) {
	raw := "=== RUN   TestConfigLoad\n" +
		"--- PASS: TestConfigLoad (0.00s)\n" +
		"=== RUN   TestConfigSave\n" +
		"--- PASS: TestConfigSave (0.01s)\n" +
		"=== RUN   TestConfigMerge\n" +
		"--- PASS: TestConfigMerge (0.00s)\n" +
		"=== RUN   TestDBConnect\n" +
		"--- PASS: TestDBConnect (0.02s)\n" +
		"=== RUN   TestDBQuery\n" +
		"--- PASS: TestDBQuery (0.01s)\n" +
		"=== RUN   TestDBTransaction\n" +
		"--- PASS: TestDBTransaction (0.03s)\n" +
		"=== RUN   TestHandlerGetUsers\n" +
		"--- PASS: TestHandlerGetUsers (0.00s)\n" +
		"=== RUN   TestHandlerCreateUser\n" +
		"--- PASS: TestHandlerCreateUser (0.01s)\n" +
		"=== RUN   TestHandlerDeleteUser\n" +
		"--- PASS: TestHandlerDeleteUser (0.00s)\n" +
		"=== RUN   TestHandlerUpdateUser\n" +
		"--- PASS: TestHandlerUpdateUser (0.01s)\n" +
		"=== RUN   TestModelUserNew\n" +
		"--- PASS: TestModelUserNew (0.00s)\n" +
		"=== RUN   TestModelUserValidate\n" +
		"--- PASS: TestModelUserValidate (0.00s)\n" +
		"=== RUN   TestModelUserSerialize\n" +
		"--- PASS: TestModelUserSerialize (0.01s)\n" +
		"=== RUN   TestRouteHealth\n" +
		"--- PASS: TestRouteHealth (0.00s)\n" +
		"=== RUN   TestRouteIndex\n" +
		"--- PASS: TestRouteIndex (0.00s)\n" +
		"=== RUN   TestRouteNotFound\n" +
		"--- PASS: TestRouteNotFound (0.00s)\n" +
		"=== RUN   TestServiceAuth\n" +
		"--- PASS: TestServiceAuth (0.02s)\n" +
		"=== RUN   TestServiceHashPassword\n" +
		"--- PASS: TestServiceHashPassword (0.01s)\n" +
		"PASS\n" +
		"ok  \tgithub.com/example/myapp\t0.45s\n"

	got, err := filterGoTestCmd(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "all 18 tests passed") {
		t.Errorf("expected 'all 18 tests passed', got:\n%s", got)
	}
	if !strings.Contains(got, "0.45s") {
		t.Errorf("expected time in output, got:\n%s", got)
	}

	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 80.0 {
		t.Errorf("expected >=80%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
	t.Logf("token savings: %.1f%% (%d -> %d)", savings, rawTokens, filteredTokens)
	t.Logf("output: %s", got)
}

func TestFilterGoTestWithFailures(t *testing.T) {
	raw := "=== RUN   TestConfigLoad\n" +
		"--- PASS: TestConfigLoad (0.00s)\n" +
		"=== RUN   TestConfigSave\n" +
		"--- PASS: TestConfigSave (0.01s)\n" +
		"=== RUN   TestDBConnect\n" +
		"--- PASS: TestDBConnect (0.02s)\n" +
		"=== RUN   TestDBQuery\n" +
		"    db_test.go:45: expected SELECT * FROM users, got SELECT id FROM users\n" +
		"    db_test.go:46: query mismatch\n" +
		"--- FAIL: TestDBQuery (0.01s)\n" +
		"=== RUN   TestHandlerGetUsers\n" +
		"--- PASS: TestHandlerGetUsers (0.00s)\n" +
		"=== RUN   TestHandlerCreateUser\n" +
		"    handler_test.go:78: validation error: email is required\n" +
		"--- FAIL: TestHandlerCreateUser (0.01s)\n" +
		"=== RUN   TestModelUserNew\n" +
		"--- PASS: TestModelUserNew (0.00s)\n" +
		"=== RUN   TestServiceAuth\n" +
		"--- PASS: TestServiceAuth (0.02s)\n" +
		"FAIL\n" +
		"FAIL\tgithub.com/example/myapp\t0.32s\n"

	got, err := filterGoTestCmd(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Failures preserved
	if !strings.Contains(got, "TestDBQuery") {
		t.Errorf("expected TestDBQuery failure, got:\n%s", got)
	}
	if !strings.Contains(got, "TestHandlerCreateUser") {
		t.Errorf("expected TestHandlerCreateUser failure, got:\n%s", got)
	}
	if !strings.Contains(got, "email is required") {
		t.Errorf("expected error message preserved, got:\n%s", got)
	}

	// Summary
	if !strings.Contains(got, "6 passed") {
		t.Errorf("expected '6 passed' in summary, got:\n%s", got)
	}
	if !strings.Contains(got, "2 failed") {
		t.Errorf("expected '2 failed' in summary, got:\n%s", got)
	}

	// Passing test names should NOT appear
	if strings.Contains(got, "TestConfigLoad") {
		t.Errorf("expected passing test names stripped, got:\n%s", got)
	}

	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	t.Logf("token savings: %.1f%% (%d -> %d)", savings, rawTokens, filteredTokens)
	t.Logf("output:\n%s", got)
}

func TestFilterGoTestEmpty(t *testing.T) {
	got, err := filterGoTestCmd("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty output, got %q", got)
	}
}
