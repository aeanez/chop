package filters

import (
	"strings"
	"testing"
)

func TestNpmListBasic(t *testing.T) {
	raw := `my-app@1.0.0 /home/user/my-app
+-- express@4.18.2
|   +-- accepts@1.3.8
|   |   +-- mime-types@2.1.35
|   |   |   +-- mime-db@1.52.0
|   |   +-- negotiator@0.6.3
|   +-- body-parser@1.20.2
|   |   +-- bytes@3.1.2
|   |   +-- content-type@1.0.5
|   |   +-- depd@2.0.0
|   |   +-- destroy@1.2.0
|   |   +-- http-errors@2.0.0
|   |   +-- on-finished@2.4.1
|   |   +-- qs@6.11.0
|   |   +-- raw-body@2.5.2
|   |   +-- type-is@1.6.18
|   |   +-- unpipe@1.0.0
|   +-- cookie@0.5.0
|   +-- debug@2.6.9
|   +-- finalhandler@1.2.0
|   +-- merge-descriptors@1.0.1
|   +-- methods@1.1.2
|   +-- path-to-regexp@0.1.7
|   +-- serve-static@1.15.0
+-- lodash@4.17.21
+-- axios@1.6.2
|   +-- follow-redirects@1.15.3
|   +-- form-data@4.0.0
|   |   +-- asynckit@0.4.0
|   |   +-- combined-stream@1.0.8
|   |   +-- mime-types@2.1.35
|   +-- proxy-from-env@1.1.0
+-- dotenv@16.3.1
+-- winston@3.11.0
|   +-- async@3.2.5
|   +-- is-stream@2.0.1
|   +-- logform@2.6.0
|   +-- one-time@1.0.0
|   +-- readable-stream@3.6.2
|   +-- stack-trace@0.0.10
|   +-- triple-beam@1.4.1
|   +-- winston-transport@4.6.0
+-- cors@2.8.5
+-- helmet@7.1.0
+-- morgan@1.10.0
|   +-- basic-auth@2.0.1
|   +-- debug@2.6.9
|   +-- depd@2.0.0
|   +-- on-finished@2.4.1
|   +-- on-headers@1.0.2
+-- uuid@9.0.0
`
	got, err := filterNpmList(raw)
	if err != nil {
		t.Fatal(err)
	}

	// Should have top-level deps
	if !strings.Contains(got, "express@4.18.2") {
		t.Errorf("expected express, got: %s", got)
	}
	if !strings.Contains(got, "lodash@4.17.21") {
		t.Errorf("expected lodash, got: %s", got)
	}
	if !strings.Contains(got, "axios@1.6.2") {
		t.Errorf("expected axios, got: %s", got)
	}

	// Should NOT have transitive deps
	if strings.Contains(got, "accepts@") {
		t.Errorf("should not include transitive dep accepts, got: %s", got)
	}
	if strings.Contains(got, "mime-types@") {
		t.Errorf("should not include transitive dep mime-types, got: %s", got)
	}

	rawTokens := countTokens(raw)
	filteredTokens := countTokens(got)
	savings := 100.0 - (float64(filteredTokens) / float64(rawTokens) * 100.0)
	if savings < 60.0 {
		t.Errorf("expected >=60%% savings, got %.1f%%", savings)
	}
	t.Logf("token savings: %.1f%% (%d -> %d)", savings, rawTokens, filteredTokens)
	t.Logf("output:\n%s", got)
}

func TestNpmListCountsDeps(t *testing.T) {
	raw := `my-app@1.0.0 /home/user/my-app
+-- express@4.18.2
|   +-- accepts@1.3.8
|   +-- body-parser@1.20.2
+-- lodash@4.17.21
` + "`" + `-- dotenv@16.3.1
`
	got, err := filterNpmList(raw)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, "express@4.18.2") {
		t.Errorf("expected express, got: %s", got)
	}
	if !strings.Contains(got, "dotenv@16.3.1") {
		t.Errorf("expected dotenv, got: %s", got)
	}
	t.Logf("output:\n%s", got)
}
