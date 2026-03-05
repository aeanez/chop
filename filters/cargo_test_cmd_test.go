package filters

import (
	"strings"
	"testing"
)

func TestFilterCargoTestAllPassing(t *testing.T) {
	raw := "   Compiling myapp v0.1.0 (/home/user/myapp)\n" +
		"    Finished test [unoptimized + debuginfo] target(s) in 2.34s\n" +
		"     Running unittests src/lib.rs (target/debug/deps/myapp-abc123)\n" +
		"\n" +
		"running 22 tests\n" +
		"test config::tests::test_default_config ... ok\n" +
		"test config::tests::test_load_config ... ok\n" +
		"test config::tests::test_merge_config ... ok\n" +
		"test db::tests::test_connection ... ok\n" +
		"test db::tests::test_query ... ok\n" +
		"test db::tests::test_transaction ... ok\n" +
		"test handlers::tests::test_get_users ... ok\n" +
		"test handlers::tests::test_create_user ... ok\n" +
		"test handlers::tests::test_delete_user ... ok\n" +
		"test handlers::tests::test_update_user ... ok\n" +
		"test models::tests::test_user_new ... ok\n" +
		"test models::tests::test_user_validate ... ok\n" +
		"test models::tests::test_user_serialize ... ok\n" +
		"test models::tests::test_role_default ... ok\n" +
		"test routes::tests::test_health ... ok\n" +
		"test routes::tests::test_index ... ok\n" +
		"test routes::tests::test_not_found ... ok\n" +
		"test services::tests::test_auth ... ok\n" +
		"test services::tests::test_hash_password ... ok\n" +
		"test services::tests::test_verify_token ... ok\n" +
		"test utils::tests::test_slugify ... ok\n" +
		"test utils::tests::test_truncate ... ok\n" +
		"\n" +
		"test result: ok. 22 passed; 0 failed; 0 ignored; 0 measured; 0 filtered out; finished in 0.45s\n" +
		"\n" +
		"   Doc-tests myapp\n" +
		"\n" +
		"running 3 tests\n" +
		"test src/lib.rs - example1 (line 12) ... ok\n" +
		"test src/lib.rs - example2 (line 24) ... ok\n" +
		"test src/lib.rs - example3 (line 36) ... ok\n" +
		"\n" +
		"test result: ok. 3 passed; 0 failed; 0 ignored; 0 measured; 0 filtered out; finished in 0.82s"

	got, err := filterCargoTestCmd(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "all 25 tests passed") {
		t.Errorf("expected 'all 25 tests passed', got:\n%s", got)
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

func TestFilterCargoTestWithFailures(t *testing.T) {
	raw := "   Compiling myapp v0.1.0 (/home/user/myapp)\n" +
		"    Finished test [unoptimized + debuginfo] target(s) in 1.87s\n" +
		"     Running unittests src/lib.rs (target/debug/deps/myapp-abc123)\n" +
		"\n" +
		"running 10 tests\n" +
		"test config::tests::test_default_config ... ok\n" +
		"test config::tests::test_load_config ... ok\n" +
		"test db::tests::test_connection ... ok\n" +
		"test db::tests::test_query ... FAILED\n" +
		"test handlers::tests::test_get_users ... ok\n" +
		"test handlers::tests::test_create_user ... FAILED\n" +
		"test models::tests::test_user_new ... ok\n" +
		"test models::tests::test_user_validate ... ok\n" +
		"test routes::tests::test_health ... ok\n" +
		"test services::tests::test_auth ... ok\n" +
		"\n" +
		"failures:\n" +
		"\n" +
		"---- db::tests::test_query stdout ----\n" +
		"thread 'db::tests::test_query' panicked at 'assertion failed: (left == right)\n" +
		"  left: \"SELECT * FROM users\",\n" +
		" right: \"SELECT id FROM users\"', src/db.rs:45:9\n" +
		"\n" +
		"---- handlers::tests::test_create_user stdout ----\n" +
		"thread 'handlers::tests::test_create_user' panicked at 'called Result::unwrap() on an Err value: ValidationError(\"email is required\")', src/handlers.rs:78:14\n" +
		"\n" +
		"failures:\n" +
		"    db::tests::test_query\n" +
		"    handlers::tests::test_create_user\n" +
		"\n" +
		"test result: FAILED. 8 passed; 2 failed; 0 ignored; 0 measured; 0 filtered out; finished in 0.32s"

	got, err := filterCargoTestCmd(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Failures preserved
	if !strings.Contains(got, "test_query") {
		t.Errorf("expected test_query failure, got:\n%s", got)
	}
	if !strings.Contains(got, "test_create_user") {
		t.Errorf("expected test_create_user failure, got:\n%s", got)
	}
	if !strings.Contains(got, "assertion failed") {
		t.Errorf("expected assertion message preserved, got:\n%s", got)
	}
	if !strings.Contains(got, "email is required") {
		t.Errorf("expected error message preserved, got:\n%s", got)
	}

	// Summary
	if !strings.Contains(got, "8 passed") {
		t.Errorf("expected '8 passed' in summary, got:\n%s", got)
	}
	if !strings.Contains(got, "2 failed") {
		t.Errorf("expected '2 failed' in summary, got:\n%s", got)
	}

	// Passing test names should NOT appear
	if strings.Contains(got, "test_default_config") {
		t.Errorf("expected passing test names stripped, got:\n%s", got)
	}

	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	t.Logf("token savings: %.1f%% (%d -> %d)", savings, rawTokens, filteredTokens)
	t.Logf("output:\n%s", got)
}

func TestFilterCargoTestEmpty(t *testing.T) {
	got, err := filterCargoTestCmd("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty output, got %q", got)
	}
}
