package filters

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var (
	// TSC error line: "src/app.ts(12,5): error TS2322: Type 'string' is not assignable to type 'number'."
	reTscError = regexp.MustCompile(`^(.+?)\((\d+),\d+\):\s*error\s+(TS\d+):\s*(.+)`)
	// TSC "Found N errors in N files." summary
	reTscFoundErrors = regexp.MustCompile(`(?i)^Found\s+\d+\s+errors?\s+in\s+\d+\s+files?`)
	// TSC "Found N errors." (single-file variant)
	reTscFoundErrorsSingle = regexp.MustCompile(`(?i)^Found\s+\d+\s+errors?\.`)
	// Tilde underline lines: "     ~~~~~"
	reTscTilde = regexp.MustCompile(`^\s*~+\s*$`)
)

func filterTsc(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "no errors", nil
	}
	if !looksLikeTscOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	type tscError struct {
		file string
		line string
		code string
		msg  string
	}

	var errors []tscError

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		// Skip tilde underlines
		if reTscTilde.MatchString(trimmed) {
			continue
		}

		// Skip "Found N errors" summary (we generate our own)
		if reTscFoundErrors.MatchString(trimmed) || reTscFoundErrorsSingle.MatchString(trimmed) {
			continue
		}

		// Parse error lines
		if m := reTscError.FindStringSubmatch(line); m != nil {
			errors = append(errors, tscError{
				file: m[1],
				line: m[2],
				code: m[3],
				msg:  m[4],
			})
			continue
		}
	}

	if len(errors) == 0 {
		// Check if there's any non-empty content we couldn't parse
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || reTscTilde.MatchString(trimmed) {
				continue
			}
			if reTscFoundErrors.MatchString(trimmed) || reTscFoundErrorsSingle.MatchString(trimmed) {
				continue
			}
			// Has unparseable content, return raw
			return raw, nil
		}
		return "no errors", nil
	}

	// Group by error code
	codeMap := make(map[string][]string) // code -> []file:line
	codeMsg := make(map[string]string)   // code -> first message
	fileSet := make(map[string]bool)

	for _, e := range errors {
		loc := fmt.Sprintf("%s:%s", e.file, e.line)
		codeMap[e.code] = append(codeMap[e.code], loc)
		if _, ok := codeMsg[e.code]; !ok {
			codeMsg[e.code] = e.msg
		}
		fileSet[e.file] = true
	}

	// Sort codes for deterministic output
	codes := make([]string, 0, len(codeMap))
	for c := range codeMap {
		codes = append(codes, c)
	}
	sort.Strings(codes)

	var out []string
	for _, code := range codes {
		locs := codeMap[code]
		out = append(out, fmt.Sprintf("%s (%d): %s", code, len(locs), codeMsg[code]))
		out = append(out, fmt.Sprintf("  %s", strings.Join(locs, ", ")))
	}

	out = append(out, "")
	out = append(out, fmt.Sprintf("%d errors in %d files", len(errors), len(fileSet)))

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}
