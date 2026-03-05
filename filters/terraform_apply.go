package filters

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// Matches "aws_instance.web: Creating..." or "aws_instance.web: Modifying..." etc.
	reTfApplyAction = regexp.MustCompile(`^(\S+):\s+(Creating|Modifying|Destroying|Reading)\.\.\.\s*$`)
	// Matches "aws_instance.web: Creation complete after 1m2s [id=i-123]"
	reTfApplyComplete = regexp.MustCompile(`^(\S+):\s+(Creation|Modifications|Destruction|Read)\s+complete\s+after\s+(\S+)`)
	// Matches "aws_instance.web: Still creating... [1m10s elapsed]"
	reTfApplyStill = regexp.MustCompile(`(?i)still\s+(creating|modifying|destroying|reading)`)
	// Matches "Apply complete! Resources: 2 added, 1 changed, 0 destroyed."
	reTfApplySummary = regexp.MustCompile(`Apply complete!\s+Resources:\s+(\d+)\s+added,\s+(\d+)\s+changed,\s+(\d+)\s+destroyed`)
	// Matches error lines in apply
	reTfApplyError = regexp.MustCompile(`(?i)^(error|warning)\b`)
	reTfApplyErrorColon = regexp.MustCompile(`(?i)error:`)
	// Matches "Destroy complete! Resources: N destroyed."
	reTfDestroySummary = regexp.MustCompile(`Destroy complete!\s+Resources:\s+(\d+)\s+destroyed`)
)

func filterTerraformApply(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeTerraformApplyOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	var (
		resources  []string
		summary    string
		errors     []string
		completed  = make(map[string]string) // resource -> "action (timing)"
		inProgress = make(map[string]string) // resource -> action
	)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip "Still creating..." progress lines
		if reTfApplyStill.MatchString(trimmed) {
			continue
		}

		// Capture apply summary
		if m := reTfApplySummary.FindStringSubmatch(trimmed); m != nil {
			summary = fmt.Sprintf("Apply complete! Resources: %s added, %s changed, %s destroyed", m[1], m[2], m[3])
			continue
		}

		// Capture destroy summary
		if m := reTfDestroySummary.FindStringSubmatch(trimmed); m != nil {
			summary = fmt.Sprintf("Destroy complete! Resources: %s destroyed", m[1])
			continue
		}

		// Capture action start
		if m := reTfApplyAction.FindStringSubmatch(trimmed); m != nil {
			inProgress[m[1]] = strings.ToLower(m[2])
			continue
		}

		// Capture completion with timing
		if m := reTfApplyComplete.FindStringSubmatch(trimmed); m != nil {
			resource := m[1]
			action := mapTfApplyAction(m[2])
			timing := m[3]
			completed[resource] = fmt.Sprintf("%s (%s)", action, timing)
			delete(inProgress, resource)
			continue
		}

		// Capture errors
		if reTfApplyError.MatchString(trimmed) || reTfApplyErrorColon.MatchString(trimmed) {
			errors = append(errors, trimmed)
		}
	}

	// Build resource lines from completed actions
	for resource, info := range completed {
		resources = append(resources, fmt.Sprintf("%s: %s", resource, info))
	}

	// Add in-progress resources that never completed (likely errors)
	for resource, action := range inProgress {
		if _, done := completed[resource]; !done {
			resources = append(resources, fmt.Sprintf("%s: %s (incomplete)", resource, action))
		}
	}

	var out []string

	for _, r := range resources {
		out = append(out, r)
	}

	if len(errors) > 0 {
		out = append(out, "")
		for _, e := range errors {
			out = append(out, e)
		}
	}

	if summary != "" {
		out = append(out, "")
		out = append(out, summary)
	}

	if len(out) == 0 {
		return raw, nil
	}

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}

func mapTfApplyAction(completion string) string {
	lower := strings.ToLower(completion)
	switch {
	case strings.Contains(lower, "creation"):
		return "created"
	case strings.Contains(lower, "modification"):
		return "updated"
	case strings.Contains(lower, "destruction"):
		return "destroyed"
	case strings.Contains(lower, "read"):
		return "read"
	default:
		return lower
	}
}
