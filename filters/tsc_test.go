package filters

import (
	"strings"
	"testing"
)

func TestFilterTscErrors(t *testing.T) {
	raw := "src/components/dashboard/widgets/UserProfile.tsx(12,5): error TS2322: Type 'string' is not assignable to type 'number'.\n" +
		"\n" +
		"  12   const age: number = userName;\n" +
		"       ~~~\n" +
		"\n" +
		"  The expected type comes from property 'age' which is declared here on type 'UserProps'\n" +
		"\n" +
		"src/components/dashboard/widgets/UserProfile.tsx(24,10): error TS2322: Type 'string' is not assignable to type 'number'.\n" +
		"\n" +
		"  24   const count: number = label;\n" +
		"       ~~~~~\n" +
		"\n" +
		"  The expected type comes from property 'count' which is declared here on type 'UserProps'\n" +
		"\n" +
		"src/components/dashboard/widgets/StatsPanel.tsx(8,3): error TS2322: Type 'string' is not assignable to type 'number'.\n" +
		"\n" +
		"  8   const value: number = text;\n" +
		"      ~~~~~\n" +
		"\n" +
		"  The expected type comes from property 'value' which is declared here on type 'StatsProps'\n" +
		"\n" +
		"src/components/forms/CreateUserForm.tsx(15,7): error TS2345: Argument of type 'string' is not assignable to parameter of type 'number'.\n" +
		"\n" +
		"  15   processAge(nameField);\n" +
		"       ~~~~~~~~~~~~~~~~~~~~~\n" +
		"\n" +
		"src/components/forms/CreateUserForm.tsx(32,12): error TS2345: Argument of type 'string' is not assignable to parameter of type 'number'.\n" +
		"\n" +
		"  32   validateCount(inputStr);\n" +
		"       ~~~~~~~~~~~~~~~~~~~~~~~\n" +
		"\n" +
		"src/services/api/endpoints/users.ts(45,8): error TS2339: Property 'data' does not exist on type 'Response'.\n" +
		"\n" +
		"  45   return response.data;\n" +
		"                       ~~~~\n" +
		"\n" +
		"src/services/api/endpoints/users.ts(67,14): error TS2339: Property 'data' does not exist on type 'Response'.\n" +
		"\n" +
		"  67   const result = response.data;\n" +
		"                               ~~~~\n" +
		"\n" +
		"src/services/api/endpoints/posts.ts(89,5): error TS2339: Property 'data' does not exist on type 'Response'.\n" +
		"\n" +
		"  89   response.data;\n" +
		"                ~~~~\n" +
		"\n" +
		"src/models/entities/User.ts(10,3): error TS2564: Property 'name' has no initializer and is not definitely assigned.\n" +
		"\n" +
		"  10   name: string;\n" +
		"       ~~~~\n" +
		"\n" +
		"src/models/entities/User.ts(11,3): error TS2564: Property 'email' has no initializer and is not definitely assigned.\n" +
		"\n" +
		"  11   email: string;\n" +
		"       ~~~~~\n" +
		"\n" +
		"src/routes/pages/index.tsx(5,22): error TS7016: Could not find a declaration file for module 'react-router'.\n" +
		"\n" +
		"  5   import { Router } from 'react-router';\n" +
		"                             ~~~~~~~~~~~~~~\n" +
		"\n" +
		"  Try `npm i --save-dev @types/react-router` if it exists or add a new declaration (.d.ts) file containing `declare module 'react-router';`\n" +
		"\n" +
		"Found 11 errors in 6 files.\n"

	got, err := filterTsc(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Grouped by code
	if !strings.Contains(got, "TS2322 (3)") {
		t.Errorf("expected 'TS2322 (3)', got:\n%s", got)
	}
	if !strings.Contains(got, "TS2345 (2)") {
		t.Errorf("expected 'TS2345 (2)', got:\n%s", got)
	}
	if !strings.Contains(got, "TS2339 (3)") {
		t.Errorf("expected 'TS2339 (3)', got:\n%s", got)
	}

	// Summary
	if !strings.Contains(got, "11 errors in 7 files") {
		t.Errorf("expected '11 errors in 7 files', got:\n%s", got)
	}

	// Tilde lines stripped
	if strings.Contains(got, "~~~~~~") {
		t.Errorf("expected tilde lines stripped, got:\n%s", got)
	}

	// "Found N errors" replaced with compact summary
	if strings.Contains(got, "Found 11") {
		t.Errorf("expected verbose summary stripped, got:\n%s", got)
	}

	// Source code snippets stripped
	if strings.Contains(got, "const age") {
		t.Errorf("expected source code stripped, got:\n%s", got)
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

func TestFilterTscClean(t *testing.T) {
	got, err := filterTsc("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "no errors" {
		t.Errorf("expected 'no errors', got %q", got)
	}
}
