package filters

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// Go build/vet error: "file.go:12:5: error message"
	reGoBuildError = regexp.MustCompile(`^(.+\.go):(\d+):(\d+):\s*(.+)`)
	// Go build error without column: "file.go:12: error message"
	reGoBuildErrorNoCol = regexp.MustCompile(`^(.+\.go):(\d+):\s*(.+)`)
	// "# package/path" header
	reGoBuildPkgHeader = regexp.MustCompile(`^#\s+\S+`)
)

func filterGoBuild(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "build ok", nil
	}

	lines := strings.Split(raw, "\n")

	type buildError struct {
		file    string
		line    string
		col     string
		message string
	}

	var (
		errors []buildError
		seen   = make(map[string]bool) // dedup by file:line:message
	)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		// Skip package header lines
		if reGoBuildPkgHeader.MatchString(trimmed) {
			continue
		}

		// file.go:12:5: message
		if m := reGoBuildError.FindStringSubmatch(trimmed); m != nil {
			key := m[1] + ":" + m[2] + ":" + m[4]
			if !seen[key] {
				seen[key] = true
				errors = append(errors, buildError{
					file:    m[1],
					line:    m[2],
					col:     m[3],
					message: m[4],
				})
			}
			continue
		}

		// file.go:12: message (no column)
		if m := reGoBuildErrorNoCol.FindStringSubmatch(trimmed); m != nil {
			key := m[1] + ":" + m[2] + ":" + m[3]
			if !seen[key] {
				seen[key] = true
				errors = append(errors, buildError{
					file:    m[1],
					line:    m[2],
					message: m[3],
				})
			}
			continue
		}
	}

	if len(errors) == 0 {
		return "build ok", nil
	}

	var out []string
	for _, e := range errors {
		out = append(out, fmt.Sprintf("%s:%s: %s", e.file, e.line, e.message))
	}
	out = append(out, fmt.Sprintf("\n%d error(s)", len(errors)))

	return strings.Join(out, "\n"), nil
}
