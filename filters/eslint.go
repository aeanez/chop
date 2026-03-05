package filters

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var (
	// ESLint problem line: "  12:5  error  Unexpected var  no-var"
	// or: "  12:5  warning  Missing semicolon  semi"
	reEslintProblem = regexp.MustCompile(`^\s*(\d+):(\d+)\s+(error|warning)\s+(.+?)\s{2,}(\S+)\s*$`)
	// ESLint file header: "/path/to/file.ts" or "src/file.ts"
	reEslintFileHeader = regexp.MustCompile(`^(\S.*\.\w+)\s*$`)
	// ESLint summary: "N problems (M errors, K warnings)"
	reEslintSummary = regexp.MustCompile(`(\d+)\s+problems?\s+\((\d+)\s+errors?,\s*(\d+)\s+warnings?\)`)
	// ESLint fixable: "N errors and M warnings potentially fixable"
	reEslintFixable = regexp.MustCompile(`(\d+)\s+errors?\s+and\s+(\d+)\s+warnings?\s+potentially\s+fixable`)
	// Biome summary: similar patterns
	// Source code snippet lines (indented with pipe or just code)
	reEslintSourceLine = regexp.MustCompile(`^\s*\d+\s*\|`)
	// Caret indicator lines
	reEslintCaret = regexp.MustCompile(`^\s*\^+\s*$`)
)

func filterEslint(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "no problems", nil
	}

	lines := strings.Split(raw, "\n")

	type eslintProblem struct {
		file    string
		line    string
		level   string // "error" or "warning"
		message string
		rule    string
	}

	var (
		problems    []eslintProblem
		currentFile string
		fixableMsg  string
		errorCount  int
		warnCount   int
	)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		// Skip source code snippets and carets
		if reEslintSourceLine.MatchString(line) || reEslintCaret.MatchString(trimmed) {
			continue
		}

		// Check for fixable message
		if m := reEslintFixable.FindStringSubmatch(trimmed); m != nil {
			fixableMsg = trimmed
			continue
		}

		// Skip summary line (we generate our own)
		if reEslintSummary.MatchString(trimmed) {
			continue
		}

		// Problem line
		if m := reEslintProblem.FindStringSubmatch(line); m != nil {
			p := eslintProblem{
				file:    currentFile,
				line:    m[1],
				level:   m[3],
				message: m[4],
				rule:    m[5],
			}
			problems = append(problems, p)
			if m[3] == "error" {
				errorCount++
			} else {
				warnCount++
			}
			continue
		}

		// File header (non-indented line ending in file extension)
		if reEslintFileHeader.MatchString(trimmed) {
			currentFile = trimmed
			continue
		}
	}

	if len(problems) == 0 {
		return "no problems", nil
	}

	// Group by rule
	ruleMap := make(map[string][]string) // rule -> []file:line
	ruleLevel := make(map[string]string) // rule -> level
	ruleMsg := make(map[string]string)   // rule -> first message

	for _, p := range problems {
		loc := fmt.Sprintf("%s:%s", p.file, p.line)
		ruleMap[p.rule] = append(ruleMap[p.rule], loc)
		if _, ok := ruleLevel[p.rule]; !ok {
			ruleLevel[p.rule] = p.level
			ruleMsg[p.rule] = p.message
		}
	}

	// Sort rules for deterministic output
	rules := make([]string, 0, len(ruleMap))
	for r := range ruleMap {
		rules = append(rules, r)
	}
	sort.Strings(rules)

	var out []string
	for _, rule := range rules {
		locs := ruleMap[rule]
		out = append(out, fmt.Sprintf("%s (%d): %s", rule, len(locs), strings.Join(locs, ", ")))
	}

	out = append(out, "")
	out = append(out, fmt.Sprintf("%d problems (%d errors, %d warnings)", len(problems), errorCount, warnCount))

	if fixableMsg != "" {
		out = append(out, fixableMsg)
	}

	return strings.Join(out, "\n"), nil
}
