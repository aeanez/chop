package filters

import (
	"fmt"
	"regexp"
	"strings"
)

var reListDep = regexp.MustCompile(`^[a-z@].*@\d|^[+`+"`"+`|\\/ ]+[a-z@].*@\d`)
var reTopLevelDep = regexp.MustCompile(`^[+`+"`"+`|\\]+--\s+(.+@\S+)`)
var reDirectDep = regexp.MustCompile(`^[+`+"`"+`]--\s+(.+@\S+)`)

func filterNpmList(raw string) (string, error) {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	if len(lines) == 0 {
		return raw, nil
	}

	var topLevel []string
	totalCount := 0

	for _, line := range lines {
		// Count all packages (any line with name@version pattern)
		if strings.Contains(line, "@") && reListDep.MatchString(strings.TrimSpace(line)) {
			totalCount++
		}

		// Top-level deps are prefixed with +-- or `-- (single level of nesting)
		if m := reDirectDep.FindStringSubmatch(line); m != nil {
			topLevel = append(topLevel, m[1])
		}
	}

	if len(topLevel) == 0 {
		// Fallback: try to extract from simpler format (npm ls --depth=0)
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			// Lines with +-- or `-- at start
			if m := reTopLevelDep.FindStringSubmatch(line); m != nil {
				topLevel = append(topLevel, m[1])
			} else if strings.Contains(trimmed, "@") && !strings.HasPrefix(trimmed, " ") && !strings.Contains(trimmed, "──") {
				// Direct package@version lines
				if idx := strings.Index(trimmed, " "); idx == -1 {
					topLevel = append(topLevel, trimmed)
				}
			}
		}
	}

	if len(topLevel) == 0 {
		return raw, nil
	}

	var out strings.Builder
	for _, dep := range topLevel {
		fmt.Fprintln(&out, dep)
	}
	if totalCount > len(topLevel) {
		fmt.Fprintf(&out, "\n%d direct deps, %d total packages", len(topLevel), totalCount)
	} else {
		fmt.Fprintf(&out, "\n%d packages", len(topLevel))
	}

	return strings.TrimSpace(out.String()), nil
}
