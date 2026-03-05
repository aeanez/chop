package filters

import (
	"regexp"
	"strings"
)

var (
	// Matches "- Installing hashicorp/aws v5.31.0..."
	reTfInitInstalling = regexp.MustCompile(`(?i)-\s+installing\s+(\S+)\s+(v\S+)`)
	// Matches "- Installed hashicorp/aws v5.31.0 (signed by ...)"
	reTfInitInstalled = regexp.MustCompile(`(?i)-\s+installed\s+(\S+)\s+(v\S+)`)
	// Matches "Terraform has been successfully initialized!"
	reTfInitSuccess = regexp.MustCompile(`(?i)terraform has been successfully initialized`)
	// Matches error lines
	reTfInitError = regexp.MustCompile(`(?i)^(error|warning)\b`)
	reTfInitErrorColon = regexp.MustCompile(`(?i)error:`)
	// Matches "Reusing previous version" noise
	reTfInitReusing = regexp.MustCompile(`(?i)reusing previous version`)
	// Matches downloading/progress lines
	reTfInitDownloading = regexp.MustCompile(`(?i)(downloading|finding|initializing)`)
	// Matches partner/signing info
	reTfInitPartner = regexp.MustCompile(`(?i)(partner|signed by|signature)`)
)

func filterTerraformInit(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}

	lines := strings.Split(raw, "\n")

	var (
		providers = make(map[string]string) // provider -> version (dedup)
		success   bool
		errors    []string
	)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Capture installed providers (prefer "Installed" over "Installing")
		if m := reTfInitInstalled.FindStringSubmatch(trimmed); m != nil {
			providers[m[1]] = m[2]
			continue
		}

		// Capture installing providers (only if not already installed)
		if m := reTfInitInstalling.FindStringSubmatch(trimmed); m != nil {
			if _, exists := providers[m[1]]; !exists {
				providers[m[1]] = m[2]
			}
			continue
		}

		// Capture success
		if reTfInitSuccess.MatchString(trimmed) {
			success = true
			continue
		}

		// Capture errors
		if reTfInitError.MatchString(trimmed) || reTfInitErrorColon.MatchString(trimmed) {
			errors = append(errors, trimmed)
		}
	}

	var out []string

	if len(providers) > 0 {
		for provider, version := range providers {
			out = append(out, provider+" "+version)
		}
	}

	if len(errors) > 0 {
		out = append(out, "")
		for _, e := range errors {
			out = append(out, e)
		}
	}

	if success {
		if len(out) > 0 {
			out = append(out, "")
		}
		out = append(out, "Terraform has been successfully initialized!")
	}

	if len(out) == 0 {
		return raw, nil
	}

	return strings.Join(out, "\n"), nil
}
