package filters

import (
	"fmt"
	"strings"
)

func filterGitLog(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}

	lines := strings.Split(trimmed, "\n")

	// Detect oneline format: no line starts with "commit "
	isVerbose := false
	for _, line := range lines {
		if strings.HasPrefix(line, "commit ") && len(line) >= 47 {
			isVerbose = true
			break
		}
	}
	if !isVerbose {
		return trimmed, nil
	}

	// Parse verbose format into condensed one-line entries
	type entry struct {
		hash    string
		message string
	}
	var entries []entry
	var currentHash string
	var msgLines []string

	flush := func() {
		if currentHash != "" {
			// Use only the first non-empty message line (subject line)
			msg := ""
			for _, ml := range msgLines {
				if ml != "" {
					msg = ml
					break
				}
			}
			if msg == "" {
				msg = "(no message)"
			}
			short := currentHash
			if len(short) > 7 {
				short = short[:7]
			}
			entries = append(entries, entry{hash: short, message: msg})
		}
		currentHash = ""
		msgLines = nil
	}

	inHeader := true
	for _, line := range lines {
		if strings.HasPrefix(line, "commit ") {
			flush()
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				currentHash = parts[1]
			}
			inHeader = true
			continue
		}
		if inHeader {
			if strings.HasPrefix(line, "Author:") || strings.HasPrefix(line, "Date:") ||
				strings.HasPrefix(line, "Merge:") {
				continue
			}
			if line == "" {
				inHeader = false
				continue
			}
		}
		// Message body line
		if !inHeader {
			t := strings.TrimSpace(line)
			if t != "" {
				msgLines = append(msgLines, t)
			}
		}
	}
	flush()

	const maxEntries = 20
	total := len(entries)
	if total > maxEntries {
		entries = entries[:maxEntries]
	}

	var out strings.Builder
	for _, e := range entries {
		fmt.Fprintf(&out, "%s %s\n", e.hash, e.message)
	}
	if total > maxEntries {
		fmt.Fprintf(&out, "(%d more)\n", total-maxEntries)
	}

	return strings.TrimSpace(out.String()), nil
}
