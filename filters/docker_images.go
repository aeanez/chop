package filters

import (
	"fmt"
	"strings"
)

func filterDockerImages(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}

	lines := strings.Split(raw, "\n")
	if len(lines) == 0 {
		return "", nil
	}

	header := lines[0]
	repoIdx := strings.Index(header, "REPOSITORY")
	tagIdx := strings.Index(header, "TAG")
	sizeIdx := strings.Index(header, "SIZE")

	if repoIdx == -1 || tagIdx == -1 || sizeIdx == -1 {
		return raw, nil
	}

	// Find IMAGE ID column to bound TAG end
	imageIDIdx := strings.Index(header, "IMAGE ID")
	if imageIDIdx == -1 {
		imageIDIdx = sizeIdx
	}

	var tagged []string
	var noneEntries []string

	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}

		repo := extractColumn(line, repoIdx, tagIdx)
		tag := extractColumn(line, tagIdx, imageIDIdx)
		size := extractColumn(line, sizeIdx, len(line))

		if repo == "<none>" || tag == "<none>" {
			noneEntries = append(noneEntries, fmt.Sprintf("%s:%s %s", repo, tag, size))
			continue
		}

		tagged = append(tagged, fmt.Sprintf("%s:%s %s", repo, tag, size))
	}

	var result []string
	if len(tagged) > 0 {
		result = tagged
	} else {
		// Only show <none> images if that's all there is
		result = noneEntries
	}

	total := len(tagged) + len(noneEntries)
	if total == 0 {
		return "", nil
	}

	result = append(result, fmt.Sprintf("%d images total", total))
	return strings.Join(result, "\n"), nil
}
