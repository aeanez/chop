package filters

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// Matches plan summary line: "Plan: 3 to add, 1 to change, 0 to destroy."
	reTfPlanSummary = regexp.MustCompile(`Plan:\s+(\d+)\s+to add,\s+(\d+)\s+to change,\s+(\d+)\s+to destroy`)
	// Matches resource action headers like "# aws_instance.web will be created"
	reTfResourceAction = regexp.MustCompile(`^#\s+(\S+)\s+will be\s+(.+)$`)
	// Matches resource action headers like "# aws_instance.web must be replaced"
	reTfResourceMustReplace = regexp.MustCompile(`^#\s+(\S+)\s+must be\s+(.+)$`)
	// Matches changed attribute lines (with ~)
	reTfChangedAttr = regexp.MustCompile(`^\s+~\s+(.+)`)
	// Matches "No changes." or "No changes. Your infrastructure matches"
	reTfNoChanges = regexp.MustCompile(`(?i)no changes`)
	// Matches error/warning lines
	reTfPlanError = regexp.MustCompile(`(?i)^(error|warning)\b`)
	// Matches lines starting with "Error:" anywhere
	reTfErrorColon = regexp.MustCompile(`(?i)error:`)
)

func filterTerraformPlan(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeTerraformPlanOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	// Check for no changes
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if reTfNoChanges.MatchString(trimmed) && !reTfPlanError.MatchString(trimmed) {
			return "no changes", nil
		}
	}

	var (
		resources    []string
		changedAttrs []string
		currentRes   string
		summary      string
		errors       []string
		inResource   bool
	)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Capture summary
		if m := reTfPlanSummary.FindStringSubmatch(trimmed); m != nil {
			summary = fmt.Sprintf("Plan: %s to add, %s to change, %s to destroy", m[1], m[2], m[3])
			continue
		}

		// Capture resource action
		if m := reTfResourceAction.FindStringSubmatch(trimmed); m != nil {
			// Flush previous resource
			if currentRes != "" {
				entry := currentRes
				if len(changedAttrs) > 0 {
					entry += "\n" + strings.Join(changedAttrs, "\n")
				}
				resources = append(resources, entry)
				changedAttrs = nil
			}
			action := mapTfAction(m[2])
			currentRes = fmt.Sprintf("~ %s (%s)", m[1], action)
			inResource = true
			continue
		}

		if m := reTfResourceMustReplace.FindStringSubmatch(trimmed); m != nil {
			if currentRes != "" {
				entry := currentRes
				if len(changedAttrs) > 0 {
					entry += "\n" + strings.Join(changedAttrs, "\n")
				}
				resources = append(resources, entry)
				changedAttrs = nil
			}
			currentRes = fmt.Sprintf("~ %s (replace)", m[1])
			inResource = true
			continue
		}

		// Capture changed attributes within resource block
		if inResource {
			if m := reTfChangedAttr.FindStringSubmatch(line); m != nil {
				changedAttrs = append(changedAttrs, "    ~ "+strings.TrimSpace(m[1]))
				continue
			}
		}

		// Capture errors/warnings
		if reTfPlanError.MatchString(trimmed) || reTfErrorColon.MatchString(trimmed) {
			errors = append(errors, trimmed)
		}
	}

	// Flush last resource
	if currentRes != "" {
		entry := currentRes
		if len(changedAttrs) > 0 {
			entry += "\n" + strings.Join(changedAttrs, "\n")
		}
		resources = append(resources, entry)
	}

	var out []string

	if summary != "" {
		out = append(out, summary)
		out = append(out, "")
	}

	for _, r := range resources {
		out = append(out, r)
	}

	if len(errors) > 0 {
		out = append(out, "")
		for _, e := range errors {
			out = append(out, e)
		}
	}

	if len(out) == 0 {
		return raw, nil
	}

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}

func mapTfAction(desc string) string {
	lower := strings.ToLower(desc)
	switch {
	case strings.Contains(lower, "destroy"):
		return "destroy"
	case strings.Contains(lower, "creat"):
		return "create"
	case strings.Contains(lower, "update"):
		return "update"
	case strings.Contains(lower, "replace"):
		return "replace"
	case strings.Contains(lower, "read"):
		return "read"
	default:
		return strings.TrimSuffix(strings.TrimSpace(desc), ".")
	}
}
