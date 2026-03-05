package filters

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// "test result: ok. N passed; M failed; ..." or "test result: FAILED. ..."
	reCargoTestResult = regexp.MustCompile(`(?m)^test result:\s*(ok|FAILED)\.\s*(\d+)\s*passed;\s*(\d+)\s*failed;\s*(\d+)\s*ignored;`)
	// Individual test line: "test module::test_name ... ok" or "... FAILED"
	reCargoTestLine = regexp.MustCompile(`(?m)^test\s+(\S+)\s+\.\.\.\s+(ok|FAILED|ignored)`)
	// "running N tests" / "running N test"
	reCargoTestRunning = regexp.MustCompile(`(?m)^running \d+ tests?$`)
	// Doc-test header: "Doc-tests crate_name"
	reCargoDocTest = regexp.MustCompile(`(?m)^Doc-tests\s+`)
	// Compilation noise: "Compiling ...", "Downloading ...", "Finished ...", "Executable ..."
	reCargoTestNoise = regexp.MustCompile(`(?i)^\s*(Compiling|Downloading|Finished|Executable|warning:)\s+`)
	// Failure block header: "---- test_name stdout ----"
	reCargoFailureHeader = regexp.MustCompile(`^-+\s+(.+?)\s+stdout\s+-+$`)
	// "failures:" section marker
	reCargoFailuresMarker = regexp.MustCompile(`(?m)^failures:$`)
)

func filterCargoTestCmd(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}

	lines := strings.Split(raw, "\n")

	var (
		totalPassed  int
		totalFailed  int
		totalIgnored int
		resultCount  int
		failures     []string
		inFailBlock  bool
		inDocTests   bool
		docFailed    bool
	)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			if inFailBlock {
				failures = append(failures, "")
			}
			continue
		}

		// Skip compilation noise
		if reCargoTestNoise.MatchString(trimmed) {
			continue
		}

		// Skip "running N tests"
		if reCargoTestRunning.MatchString(trimmed) {
			continue
		}

		// Doc-tests section
		if reCargoDocTest.MatchString(trimmed) {
			inDocTests = true
			inFailBlock = false
			continue
		}

		// Test result summary line
		if m := reCargoTestResult.FindStringSubmatch(trimmed); m != nil {
			p := atoi(m[2])
			f := atoi(m[3])
			ig := atoi(m[4])
			totalPassed += p
			totalFailed += f
			totalIgnored += ig
			resultCount++
			if inDocTests && f > 0 {
				docFailed = true
			}
			inFailBlock = false
			inDocTests = false
			continue
		}

		// Failure block header: "---- test_name stdout ----"
		if reCargoFailureHeader.MatchString(trimmed) {
			inFailBlock = true
			failures = append(failures, trimmed)
			continue
		}

		// "failures:" marker
		if reCargoFailuresMarker.MatchString(trimmed) {
			inFailBlock = false
			continue
		}

		// Individual test line: skip passing, capture failed
		if m := reCargoTestLine.FindStringSubmatch(trimmed); m != nil {
			if m[2] == "FAILED" {
				// only add if not already in failures block
				inFailBlock = false
			}
			continue
		}

		// Collect failure context
		if inFailBlock {
			failures = append(failures, line)
			continue
		}
	}

	// Fallback if no results found
	if resultCount == 0 {
		return raw, nil
	}

	totalAll := totalPassed + totalFailed + totalIgnored

	// All passed
	if totalFailed == 0 && totalAll > 0 {
		return fmt.Sprintf("all %d tests passed", totalAll), nil
	}

	// Build failure output
	var out strings.Builder

	if len(failures) > 0 {
		for _, f := range failures {
			fmt.Fprintln(&out, f)
		}
		out.WriteString("\n")
	}

	var parts []string
	if totalPassed > 0 {
		parts = append(parts, fmt.Sprintf("%d passed", totalPassed))
	}
	if totalFailed > 0 {
		parts = append(parts, fmt.Sprintf("%d failed", totalFailed))
	}
	if totalIgnored > 0 {
		parts = append(parts, fmt.Sprintf("%d ignored", totalIgnored))
	}
	fmt.Fprintf(&out, "%s", strings.Join(parts, ", "))

	if docFailed {
		out.WriteString(" (includes doc-test failures)")
	}

	return strings.TrimSpace(out.String()), nil
}
