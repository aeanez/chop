package filters

import (
	"fmt"
	"strings"
)

func filterGitStatus(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return raw, nil
	}
	if !looksLikeGitStatusOutput(trimmed) {
		return raw, nil
	}

	lines := strings.Split(trimmed, "\n")

	var staged, modified, untracked []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "new file:"):
			staged = append(staged, strings.TrimPrefix(trimmed, "new file:"))
		case strings.HasPrefix(trimmed, "modified:"):
			modified = append(modified, strings.TrimSpace(strings.TrimPrefix(trimmed, "modified:")))
		case strings.HasPrefix(trimmed, "deleted:"):
			modified = append(modified, strings.TrimSpace(strings.TrimPrefix(trimmed, "deleted:"))+" (deleted)")
		case strings.HasPrefix(trimmed, "renamed:"):
			modified = append(modified, strings.TrimSpace(strings.TrimPrefix(trimmed, "renamed:")))
		case strings.HasPrefix(line, "??"):
			// Short format untracked
			untracked = append(untracked, strings.TrimSpace(line[2:]))
		case strings.HasPrefix(line, "\t") && !strings.Contains(line, ":"):
			// Untracked file in long format
			untracked = append(untracked, strings.TrimSpace(line))
		}
	}

	// Detect clean working tree
	if len(staged) == 0 && len(modified) == 0 && len(untracked) == 0 {
		if strings.Contains(raw, "nothing to commit") {
			return "clean", nil
		}
		// Could not parse — fallback to raw
		return raw, nil
	}

	var out strings.Builder

	if len(staged) > 0 {
		fmt.Fprintf(&out, "staged(%d): %s\n", len(staged), strings.Join(staged, ", "))
	}
	if len(modified) > 0 {
		fmt.Fprintf(&out, "modified(%d): %s\n", len(modified), strings.Join(modified, ", "))
	}
	if len(untracked) > 0 {
		fmt.Fprintf(&out, "untracked(%d): %s\n", len(untracked), strings.Join(untracked, ", "))
	}

	result := strings.TrimSpace(out.String())
	return outputSanityCheck(raw, result), nil
}
