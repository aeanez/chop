package filters

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// Cargo error: "error[E0308]: mismatched types"
	reCargoError = regexp.MustCompile(`^error(\[E\d+\])?:\s*(.+)`)
	// Cargo warning: "warning: unused variable `x`"
	reCargoWarning = regexp.MustCompile(`^warning:\s*(.+)`)
	// File location line: "  --> src/main.rs:12:5"
	reCargoLocation = regexp.MustCompile(`^\s*-->\s*(.+?):(\d+):\d+`)
	// Compilation noise: "Compiling crate v0.1.0 (/path)", "Downloading", "Finished", "Downloaded"
	reCargoBuildNoise = regexp.MustCompile(`(?i)^\s*(Compiling|Checking|Downloading|Downloaded|Finished|Packaging|Archiving|Uploading|Updating|Locking|Adding|Removing|Fresh)\s+`)
	// "warning: N warnings emitted" or "warning: `crate` (lib) generated N warnings"
	reCargoWarningSummary = regexp.MustCompile(`^warning:.*(\d+) warnings?`)
	// help/note lines: "  = help: ...", "  = note: ..."
	reCargoHelpNote = regexp.MustCompile(`^\s*=\s*(help|note):\s*`)
	// "For more information about this error..."
	reCargoMoreInfo = regexp.MustCompile(`^For more information about this error`)
	// "could not compile" line
	reCargoCouldNotCompile = regexp.MustCompile(`^error: could not compile`)
	// aborting due to
	reCargoAborting = regexp.MustCompile(`^error: aborting due to`)
)

func filterCargoBuild(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !looksLikeCargoBuildOutput(trimmed) {
		return raw, nil
	}

	raw = trimmed
	lines := strings.Split(raw, "\n")

	type diagnostic struct {
		severity string // "error" or "warning"
		code     string // e.g. "E0308" or ""
		message  string
		location string // "file.rs:12" or ""
	}

	var (
		diags       []diagnostic
		current     *diagnostic
		hasErrors   bool
		warnCount   int
		errorCount  int
	)

	flushCurrent := func() {
		if current != nil {
			diags = append(diags, *current)
			if current.severity == "error" {
				hasErrors = true
				errorCount++
			} else {
				warnCount++
			}
			current = nil
		}
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		// Skip noise
		if reCargoBuildNoise.MatchString(trimmed) {
			continue
		}

		// Skip warning summary lines
		if reCargoWarningSummary.MatchString(trimmed) {
			continue
		}

		// Skip "For more information..."
		if reCargoMoreInfo.MatchString(trimmed) {
			continue
		}

		// Skip "could not compile" and "aborting"
		if reCargoCouldNotCompile.MatchString(trimmed) {
			continue
		}
		if reCargoAborting.MatchString(trimmed) {
			continue
		}

		// Skip help/note lines
		if reCargoHelpNote.MatchString(trimmed) {
			continue
		}

		// Error line
		if m := reCargoError.FindStringSubmatch(trimmed); m != nil {
			flushCurrent()
			code := ""
			if m[1] != "" {
				code = strings.Trim(m[1], "[]")
			}
			current = &diagnostic{
				severity: "error",
				code:     code,
				message:  m[2],
			}
			continue
		}

		// Warning line
		if m := reCargoWarning.FindStringSubmatch(trimmed); m != nil {
			flushCurrent()
			current = &diagnostic{
				severity: "warning",
				message:  m[1],
			}
			continue
		}

		// Location line
		if m := reCargoLocation.FindStringSubmatch(trimmed); m != nil {
			if current != nil {
				current.location = fmt.Sprintf("%s:%s", m[1], m[2])
			}
			continue
		}
	}

	flushCurrent()

	// Fallback if nothing found
	if len(diags) == 0 {
		// Check if it was a clean build (just noise lines)
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			if reCargoBuildNoise.MatchString(trimmed) {
				continue
			}
			// Has non-noise content we didn't parse
			return raw, nil
		}
		return "built ok", nil
	}

	if !hasErrors && warnCount == 0 {
		return "built ok", nil
	}

	var out []string

	if hasErrors {
		out = append(out, "build FAILED")
	} else {
		out = append(out, fmt.Sprintf("built (%d warnings)", warnCount))
	}

	// Show errors
	var errors, warnings []string
	for _, d := range diags {
		var line string
		if d.severity == "error" {
			if d.code != "" {
				line = fmt.Sprintf("error %s: %s", d.code, d.message)
			} else {
				line = fmt.Sprintf("error: %s", d.message)
			}
			if d.location != "" {
				line += " -> " + d.location
			}
			errors = append(errors, line)
		} else {
			line = fmt.Sprintf("warning: %s", d.message)
			if d.location != "" {
				line += " -> " + d.location
			}
			warnings = append(warnings, line)
		}
	}

	if len(errors) > 0 {
		out = append(out, "")
		for _, e := range errors {
			out = append(out, "  "+e)
		}
	}

	if len(warnings) > 0 {
		out = append(out, "")
		for _, w := range warnings {
			out = append(out, "  "+w)
		}
	}

	out = append(out, "")
	out = append(out, fmt.Sprintf("%d error(s), %d warning(s)", errorCount, warnCount))

	result := strings.Join(out, "\n")
	return outputSanityCheck(raw, result), nil
}
