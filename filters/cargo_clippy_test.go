package filters

import (
	"strings"
	"testing"
)

func TestFilterCargoClippyWithWarnings(t *testing.T) {
	raw := "    Checking myapp v0.1.0 (/home/user/myapp)\n" +
		"warning: unused variable `x`\n" +
		"  --> src/main.rs:12:9\n" +
		"   |\n" +
		"12 |     let x = 42;\n" +
		"   |         ^ help: if this is intentional, prefix it with an underscore: `_x`\n" +
		"   |\n" +
		"   = note: `#[warn(unused_variables)]` on by default\n" +
		"\n" +
		"warning: unused variable `y`\n" +
		"  --> src/handlers.rs:55:9\n" +
		"   |\n" +
		"55 |     let y = get_value();\n" +
		"   |         ^ help: if this is intentional, prefix it with an underscore: `_y`\n" +
		"   |\n" +
		"   = note: `#[warn(unused_variables)]` on by default\n" +
		"\n" +
		"warning: unused import: `std::collections::HashMap`\n" +
		"  --> src/utils.rs:3:5\n" +
		"   |\n" +
		"3  | use std::collections::HashMap;\n" +
		"   |     ^^^^^^^^^^^^^^^^^^^^^^^^^\n" +
		"   |\n" +
		"   = note: `#[warn(unused_imports)]` on by default\n" +
		"\n" +
		"warning: this function has too many arguments (8/7)\n" +
		"  --> src/services.rs:15:1\n" +
		"   |\n" +
		"15 | fn create_user(a: u32, b: u32, c: u32, d: u32, e: u32, f: u32, g: u32, h: u32) {\n" +
		"   | ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^\n" +
		"   |\n" +
		"   = note: `#[warn(clippy::too_many_arguments)]` on by default\n" +
		"   = help: for further information visit https://rust-lang.github.io/rust-clippy/master/index.html#too_many_arguments\n" +
		"\n" +
		"warning: this expression creates a reference which is immediately dereferenced\n" +
		"  --> src/db.rs:22:10\n" +
		"   |\n" +
		"22 |     foo(&*bar);\n" +
		"   |         ^^^^^ help: change this to: `bar`\n" +
		"   |\n" +
		"   = note: `#[warn(clippy::needless_borrow)]` on by default\n" +
		"   = help: for further information visit https://rust-lang.github.io/rust-clippy/master/index.html#needless_borrow\n" +
		"\n" +
		"warning: this expression creates a reference which is immediately dereferenced\n" +
		"  --> src/handlers.rs:88:14\n" +
		"   |\n" +
		"88 |     process(&*input);\n" +
		"   |             ^^^^^^^ help: change this to: `input`\n" +
		"   |\n" +
		"   = help: for further information visit https://rust-lang.github.io/rust-clippy/master/index.html#needless_borrow\n" +
		"\n" +
		"warning: `myapp` (bin \"myapp\") generated 6 warnings\n" +
		"    Finished `dev` profile [unoptimized + debuginfo] target(s) in 1.23s"

	got, err := filterCargoClippy(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be grouped by rule
	if !strings.Contains(got, "Warnings:") {
		t.Errorf("expected 'Warnings:' header, got:\n%s", got)
	}

	// Check grouping: unused_variables should show 2 occurrences
	if !strings.Contains(got, "(2)") {
		t.Errorf("expected grouped count (2) for repeated lint, got:\n%s", got)
	}

	// clippy::needless_borrow should show 2 occurrences
	if !strings.Contains(got, "needless_borrow") || !strings.Contains(got, "src/db.rs:22") {
		t.Errorf("expected needless_borrow with location, got:\n%s", got)
	}

	// Summary
	if !strings.Contains(got, "6 warning(s)") {
		t.Errorf("expected '6 warning(s)' in summary, got:\n%s", got)
	}
	if !strings.Contains(got, "0 error(s)") {
		t.Errorf("expected '0 error(s)' in summary, got:\n%s", got)
	}

	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 60.0 {
		t.Errorf("expected >=60%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
	t.Logf("token savings: %.1f%% (%d -> %d)", savings, rawTokens, filteredTokens)
	t.Logf("output:\n%s", got)
}

func TestFilterCargoClippyClean(t *testing.T) {
	raw := "    Checking myapp v0.1.0 (/home/user/myapp)\n" +
		"    Finished `dev` profile [unoptimized + debuginfo] target(s) in 0.89s"

	got, err := filterCargoClippy(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "no warnings" {
		t.Errorf("expected 'no warnings', got:\n%s", got)
	}

	t.Logf("output: %s", got)
}

func TestFilterCargoClippyEmpty(t *testing.T) {
	got, err := filterCargoClippy("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty output, got %q", got)
	}
}
