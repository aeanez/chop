package filters

import (
	"strings"
)

func filterGitBranch(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}

	lines := strings.Split(trimmed, "\n")

	var current string
	var branches []string

	for _, line := range lines {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}
		if strings.HasPrefix(name, "* ") {
			current = strings.TrimPrefix(name, "* ")
		} else {
			branches = append(branches, name)
		}
	}

	var out strings.Builder
	if current != "" {
		out.WriteString("* ")
		out.WriteString(current)
		out.WriteString("\n")
	}
	for _, b := range branches {
		out.WriteString("  ")
		out.WriteString(b)
		out.WriteString("\n")
	}

	return strings.TrimSpace(out.String()), nil
}
