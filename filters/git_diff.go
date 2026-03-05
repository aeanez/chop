package filters

import (
	"fmt"
	"strings"
)

func filterGitDiff(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}

	lines := strings.Split(trimmed, "\n")

	// Short diff: pass through as-is
	if len(lines) < 10 {
		return trimmed, nil
	}

	type fileStat struct {
		name    string
		added   int
		removed int
	}
	var stats []fileStat
	var current *fileStat

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			if current != nil {
				stats = append(stats, *current)
			}
			// Extract filename from "diff --git a/foo b/foo"
			parts := strings.SplitN(line, " b/", 2)
			name := line
			if len(parts) == 2 {
				name = parts[1]
			}
			current = &fileStat{name: name}
			continue
		}
		if current == nil {
			continue
		}
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			current.added++
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			current.removed++
		}
	}
	if current != nil {
		stats = append(stats, *current)
	}

	if len(stats) == 0 {
		return trimmed, nil
	}

	var out strings.Builder
	totalAdded := 0
	totalRemoved := 0
	for _, s := range stats {
		fmt.Fprintf(&out, "%s: +%d -%d\n", s.name, s.added, s.removed)
		totalAdded += s.added
		totalRemoved += s.removed
	}
	fmt.Fprintf(&out, "%d files changed, +%d -%d", len(stats), totalAdded, totalRemoved)

	return out.String(), nil
}
