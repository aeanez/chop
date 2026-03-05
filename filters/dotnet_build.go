package filters

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// MSBuild error/warning format: file(line,col): error/warning CODE: message [project]
	reMSBuildDiag = regexp.MustCompile(`^\s*(.+?)\((\d+),?\d*\)\s*:\s*(error|warning)\s+([\w]+)\s*:\s*(.+?)(?:\s*\[.+\])?\s*$`)
	// Standalone error/warning without file location
	reMSBuildDiagNoLoc = regexp.MustCompile(`^\s*(error|warning)\s+([\w]+)\s*:\s*(.+)`)
	// Build result line
	reBuildResult = regexp.MustCompile(`(?i)^\s*Build (succeeded|FAILED)`)
	// Warning/error count summary: "N Warning(s)" / "N Error(s)"
	reBuildWarnCount  = regexp.MustCompile(`(\d+)\s+Warning\(s\)`)
	reBuildErrorCount = regexp.MustCompile(`(\d+)\s+Error\(s\)`)
	// Lines to skip
	reMSBuildNoise = regexp.MustCompile(`(?i)(^Microsoft \(R\)|^Copyright \(C\)|^\s*Build started|^\s*Restore complete|^\s*Determining projects|^\s*All projects are up-to-date|^\s*Time Elapsed|^\s*\d+ project\(s\) in solution|^\s*Nothing to do|^\s*Successfully created|^\s*Restored |^\s*NuGet |^\s*Assets file|^\s*Generating )`)
)

func filterDotnetBuild(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}

	lines := strings.Split(raw, "\n")

	var (
		errors   []string
		warnings []string
		result   string
		warnCnt  int
		errCnt   int
	)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Skip noise lines
		if reMSBuildNoise.MatchString(trimmed) {
			continue
		}

		// Build result
		if m := reBuildResult.FindStringSubmatch(trimmed); m != nil {
			if strings.EqualFold(m[1], "FAILED") {
				result = "Build FAILED"
			} else {
				result = "Build succeeded"
			}
			continue
		}

		// Warning/error count summary
		if m := reBuildWarnCount.FindStringSubmatch(trimmed); m != nil {
			warnCnt = atoi(m[1])
			continue
		}
		if m := reBuildErrorCount.FindStringSubmatch(trimmed); m != nil {
			errCnt = atoi(m[1])
			continue
		}

		// Diagnostics with file location: file(line): error/warning CODE: message
		if m := reMSBuildDiag.FindStringSubmatch(trimmed); m != nil {
			file := m[1]
			lineNo := m[2]
			severity := m[3]
			code := m[4]
			msg := strings.TrimSpace(m[5])
			compact := fmt.Sprintf("%s(%s): %s %s", file, lineNo, code, msg)
			if strings.EqualFold(severity, "error") {
				errors = append(errors, compact)
			} else {
				warnings = append(warnings, compact)
			}
			continue
		}

		// Diagnostics without file location
		if m := reMSBuildDiagNoLoc.FindStringSubmatch(trimmed); m != nil {
			severity := m[1]
			code := m[2]
			msg := strings.TrimSpace(m[3])
			compact := fmt.Sprintf("%s %s", code, msg)
			if strings.EqualFold(severity, "error") {
				errors = append(errors, compact)
			} else {
				warnings = append(warnings, compact)
			}
			continue
		}
	}

	// Fallback: if we found nothing useful, return raw
	if result == "" && len(errors) == 0 && len(warnings) == 0 {
		return raw, nil
	}

	var out []string

	if result == "" {
		if len(errors) > 0 {
			result = "Build FAILED"
		} else {
			result = "Build succeeded"
		}
	}
	out = append(out, result)

	if len(errors) > 0 {
		out = append(out, "")
		out = append(out, "Errors:")
		for _, e := range errors {
			out = append(out, "  "+e)
		}
	}

	if len(warnings) > 0 {
		out = append(out, "")
		out = append(out, "Warnings:")
		for _, w := range warnings {
			out = append(out, "  "+w)
		}
	}

	// Summary counts
	if errCnt > 0 || warnCnt > 0 {
		out = append(out, "")
		out = append(out, fmt.Sprintf("%d error(s), %d warning(s)", errCnt, warnCnt))
	} else if len(errors) > 0 || len(warnings) > 0 {
		out = append(out, "")
		out = append(out, fmt.Sprintf("%d error(s), %d warning(s)", len(errors), len(warnings)))
	}

	return strings.Join(out, "\n"), nil
}
