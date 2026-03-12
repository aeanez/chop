package filters

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	reGitStatLine    = regexp.MustCompile(`^\s+(.+?)\s*\|\s+(\d+)\s*[+\-Bb]*`)
	reGitStatSummary = regexp.MustCompile(`(\d+) files? changed(?:, (\d+) insertions?\(\+\))?(?:, (\d+) deletions?\(-\))?`)
)

func looksLikeGitStatOutput(s string) bool {
	lines := strings.SplitN(s, "\n", 10)
	for _, line := range lines {
		if reGitStatLine.MatchString(line) {
			return true
		}
	}
	return false
}

func filterGitStat(raw string) (string, error) {
	if !looksLikeGitStatOutput(raw) {
		return raw, nil
	}

	lines := strings.Split(strings.TrimSpace(raw), "\n")

	type fileStat struct {
		name    string
		changes int
	}

	var files []fileStat
	var summaryLine string

	for _, line := range lines {
		m := reGitStatLine.FindStringSubmatch(line)
		if m != nil {
			name := strings.TrimSpace(m[1])
			changes, _ := strconv.Atoi(m[2])
			files = append(files, fileStat{name, changes})
			continue
		}
		if reGitStatSummary.MatchString(line) {
			summaryLine = strings.TrimSpace(line)
		}
	}

	// Not enough files to bother compressing
	if len(files) <= 5 || summaryLine == "" {
		return raw, nil
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].changes > files[j].changes
	})

	const topN = 5
	limit := topN
	if len(files) < limit {
		limit = len(files)
	}

	topParts := make([]string, limit)
	for i := 0; i < limit; i++ {
		topParts[i] = fmt.Sprintf("%s (%d)", files[i].name, files[i].changes)
	}

	result := summaryLine + "\ntop: " + strings.Join(topParts, ", ")
	return outputSanityCheck(raw, result), nil
}
