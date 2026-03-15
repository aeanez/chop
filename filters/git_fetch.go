package filters

import (
	"strings"
)

func filterGitFetch(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}

	lines := strings.Split(trimmed, "\n")
	var out []string

	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}

		// Remote URL
		if strings.HasPrefix(t, "From ") {
			out = append(out, t)
			continue
		}

		// Ref update lines: "abc..def main -> origin/main", "* [new branch]", "* [new tag]", "- [deleted]"
		if strings.Contains(t, "->") || strings.HasPrefix(t, "* [new") || strings.HasPrefix(t, "- [deleted") || strings.HasPrefix(t, "+ [forced") {
			out = append(out, t)
			continue
		}

		// Error/hint lines
		if strings.HasPrefix(t, "error:") || strings.HasPrefix(t, "fatal:") || strings.HasPrefix(t, "hint:") {
			out = append(out, t)
			continue
		}

		// remote: lines - keep meaningful content (not progress)
		if strings.HasPrefix(t, "remote:") {
			content := strings.TrimSpace(strings.TrimPrefix(t, "remote:"))
			if content == "" ||
				strings.HasPrefix(content, "Enumerating") ||
				strings.HasPrefix(content, "Counting") ||
				strings.HasPrefix(content, "Compressing") ||
				strings.HasPrefix(content, "Total") ||
				strings.HasPrefix(content, "Resolving deltas") {
				continue
			}
			out = append(out, t)
			continue
		}

		// Skip transfer progress noise
		// Enumerating, Counting, Compressing, Unpacking, Total, Writing
		// (anything unmatched is likely progress - skip it)
	}

	if len(out) == 0 {
		return raw, nil
	}

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}