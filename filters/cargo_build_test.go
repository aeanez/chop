package filters

import (
	"strings"
	"testing"
)

func TestFilterCargoBuildSuccessWithWarnings(t *testing.T) {
	raw := "   Compiling proc-macro2 v1.0.78\n" +
		"   Compiling unicode-ident v1.0.12\n" +
		"   Compiling quote v1.0.35\n" +
		"   Compiling syn v2.0.48\n" +
		"   Compiling serde_derive v1.0.196\n" +
		"   Compiling serde v1.0.196\n" +
		"   Compiling myapp v0.1.0 (/home/user/myapp)\n" +
		"warning: unused variable `x`\n" +
		"  --> src/main.rs:12:9\n" +
		"   |\n" +
		"12 |     let x = 42;\n" +
		"   |         ^ help: if this is intentional, prefix it with an underscore: `_x`\n" +
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
		"warning: function `old_helper` is never used\n" +
		"  --> src/helpers.rs:28:4\n" +
		"   |\n" +
		"28 | fn old_helper() {\n" +
		"   |    ^^^^^^^^^^\n" +
		"   |\n" +
		"   = note: `#[warn(dead_code)]` on by default\n" +
		"\n" +
		"warning: `myapp` (bin \"myapp\") generated 3 warnings\n" +
		"    Finished `dev` profile [unoptimized + debuginfo] target(s) in 8.45s"

	got, err := filterCargoBuild(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "built (3 warnings)") {
		t.Errorf("expected 'built (3 warnings)', got:\n%s", got)
	}
	if !strings.Contains(got, "src/main.rs:12") {
		t.Errorf("expected file:line for warning, got:\n%s", got)
	}
	if !strings.Contains(got, "src/utils.rs:3") {
		t.Errorf("expected file:line for second warning, got:\n%s", got)
	}

	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 70.0 {
		t.Errorf("expected >=70%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
	t.Logf("token savings: %.1f%% (%d -> %d)", savings, rawTokens, filteredTokens)
	t.Logf("output:\n%s", got)
}

func TestFilterCargoBuildWithErrors(t *testing.T) {
	raw := "   Compiling myapp v0.1.0 (/home/user/myapp)\n" +
		"error[E0308]: mismatched types\n" +
		"  --> src/main.rs:15:20\n" +
		"   |\n" +
		"15 |     let x: u32 = \"hello\";\n" +
		"   |            ---   ^^^^^^^ expected `u32`, found `&str`\n" +
		"   |            |\n" +
		"   |            expected due to this\n" +
		"   |\n" +
		"   = note: expected type `u32`\n" +
		"              found type `&str`\n" +
		"\n" +
		"error[E0425]: cannot find value `foo` in this scope\n" +
		"  --> src/lib.rs:42:5\n" +
		"   |\n" +
		"42 |     foo\n" +
		"   |     ^^^ not found in this scope\n" +
		"\n" +
		"For more information about this error, try `rustc --explain E0308`.\n" +
		"error: could not compile `myapp` (bin \"myapp\") due to 2 previous errors"

	got, err := filterCargoBuild(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "build FAILED") {
		t.Errorf("expected 'build FAILED', got:\n%s", got)
	}
	if !strings.Contains(got, "E0308") {
		t.Errorf("expected error E0308 preserved, got:\n%s", got)
	}
	if !strings.Contains(got, "E0425") {
		t.Errorf("expected error E0425 preserved, got:\n%s", got)
	}
	if !strings.Contains(got, "src/main.rs:15") {
		t.Errorf("expected file:line preserved, got:\n%s", got)
	}
	if !strings.Contains(got, "src/lib.rs:42") {
		t.Errorf("expected file:line preserved, got:\n%s", got)
	}

	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	t.Logf("token savings: %.1f%% (%d -> %d)", savings, rawTokens, filteredTokens)
	t.Logf("output:\n%s", got)
}

func TestFilterCargoBuildClean(t *testing.T) {
	raw := "   Compiling myapp v0.1.0 (/home/user/myapp)\n" +
		"    Finished `dev` profile [unoptimized + debuginfo] target(s) in 2.15s"

	got, err := filterCargoBuild(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "built ok" {
		t.Errorf("expected 'built ok', got:\n%s", got)
	}

	t.Logf("output: %s", got)
}

func TestFilterCargoBuildEmpty(t *testing.T) {
	got, err := filterCargoBuild("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty output, got %q", got)
	}
}
