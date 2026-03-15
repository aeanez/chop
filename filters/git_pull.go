package filters

import (
	"strings"
)

func filterGitPull(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}

	// Pass through short messages unchanged
	if trimmed == "Already up to date." || trimmed == "Already up-to-date." {
		return trimmed, nil
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

		// Strategy / result lines
		if t == "Fast-forward" ||
			strings.HasPrefix(t, "Merge made by") ||
			strings.HasPrefix(t, "Successfully rebased") {
			out = append(out, t)
			continue
		}

		// Diffstat summary line: "N file(s) changed, ..."
		if strings.Contains(t, "changed") &&
			(strings.Contains(t, "insertion") || strings.Contains(t, "deletion")) {
			out = append(out, t)
			continue
		}

		// Conflict and auto-merge lines — always keep
		if strings.HasPrefix(t, "CONFLICT") || strings.HasPrefix(t, "Auto-merging") ||
			strings.HasPrefix(t, "Automatic merge failed") {
			out = append(out, t)
			continue
		}

		// Create mode / delete mode lines (new/removed files)
		if strings.HasPrefix(t, "create mode") || strings.HasPrefix(t, "delete mode") {
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

		// Skip: transfer progress (Unpacking, Enumerating, etc.)
		// Skip: "Updating abc..def" hash range line
		// Skip: per-file diffstat lines (contain " | ")
		// The summary line above captures the totals.
	}

	if len(out) == 0 {
		return raw, nil
	}

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}