package filters

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// Matches "Step N/M :" lines in docker build output
	reDockerBuildStep = regexp.MustCompile(`(?i)^step\s+(\d+)/(\d+)\s*:`)
	// Matches final "Successfully built <id>" or "Successfully tagged <tag>"
	reDockerBuiltID  = regexp.MustCompile(`(?i)^successfully built\s+(.+)`)
	reDockerTagged   = regexp.MustCompile(`(?i)^successfully tagged\s+(.+)`)
	// Matches warning/error lines
	reDockerBuildWarn = regexp.MustCompile(`(?i)(warning|error|failed|denied|unauthorized)`)
	// Matches "writing image" in buildkit output
	reDockerBuildKit = regexp.MustCompile(`(?i)#\d+\s+writing image\s+sha256:(\S+)`)
	// Matches buildkit step lines like "#5 [2/10] RUN ..."
	reBuildKitStep = regexp.MustCompile(`^#\d+\s+\[(\d+)/(\d+)\]`)
	// Matches buildkit "naming to" (final tag)
	reBuildKitNaming = regexp.MustCompile(`(?i)#\d+\s+naming to\s+(.+)`)
)

func filterDockerBuild(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}

	lines := strings.Split(raw, "\n")

	var (
		totalSteps    int
		completedStep int
		imageID       string
		imageTag      string
		warnings      []string
	)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Classic docker build: Step N/M
		if m := reDockerBuildStep.FindStringSubmatch(trimmed); m != nil {
			step := atoi(m[1])
			total := atoi(m[2])
			if total > totalSteps {
				totalSteps = total
			}
			if step > completedStep {
				completedStep = step
			}
			continue
		}

		// BuildKit step: #5 [2/10]
		if m := reBuildKitStep.FindStringSubmatch(trimmed); m != nil {
			step := atoi(m[1])
			total := atoi(m[2])
			if total > totalSteps {
				totalSteps = total
			}
			if step > completedStep {
				completedStep = step
			}
			continue
		}

		// Successfully built
		if m := reDockerBuiltID.FindStringSubmatch(trimmed); m != nil {
			imageID = strings.TrimSpace(m[1])
			continue
		}

		// Successfully tagged
		if m := reDockerTagged.FindStringSubmatch(trimmed); m != nil {
			imageTag = strings.TrimSpace(m[1])
			continue
		}

		// BuildKit: writing image
		if m := reDockerBuildKit.FindStringSubmatch(trimmed); m != nil {
			id := m[1]
			if len(id) > 12 {
				id = id[:12]
			}
			imageID = id
			continue
		}

		// BuildKit: naming to
		if m := reBuildKitNaming.FindStringSubmatch(trimmed); m != nil {
			imageTag = strings.TrimSpace(m[1])
			continue
		}

		// Warnings/errors
		if reDockerBuildWarn.MatchString(trimmed) {
			warnings = append(warnings, trimmed)
		}
	}

	var out []string

	if totalSteps > 0 {
		out = append(out, fmt.Sprintf("%d/%d steps completed", completedStep, totalSteps))
	}

	if len(warnings) > 0 {
		out = append(out, "")
		for _, w := range warnings {
			out = append(out, w)
		}
	}

	if imageTag != "" {
		out = append(out, fmt.Sprintf("image: %s", imageTag))
	} else if imageID != "" {
		out = append(out, fmt.Sprintf("image: %s", imageID))
	}

	if len(out) == 0 {
		return raw, nil
	}

	return strings.Join(out, "\n"), nil
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}
