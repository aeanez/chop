package filters

import (
	"strings"
	"testing"
)

func TestFilterTerraformInit(t *testing.T) {
	raw := `
Initializing the backend...

Successfully configured the backend "s3"! Terraform will automatically
use this backend unless the backend configuration changes.

Initializing provider plugins...
- Reusing previous version of hashicorp/aws from the dependency lock file
- Reusing previous version of hashicorp/random from the dependency lock file
- Finding hashicorp/null versions matching "~> 3.0"...
- Installing hashicorp/aws v5.31.0...
- Installed hashicorp/aws v5.31.0 (signed by HashiCorp)
- Installing hashicorp/random v3.6.0...
- Installed hashicorp/random v3.6.0 (signed by HashiCorp)
- Installing hashicorp/null v3.2.2...
- Installed hashicorp/null v3.2.2 (signed by HashiCorp)

Terraform has made some changes to the provider dependency selections recorded
in the .terraform.lock.hcl file. Review those changes and commit them to your
version control system.

Partner and community providers are signed by their developers.
If you'd like to know more about provider signing, you can read about it here:
https://www.terraform.io/docs/cli/plugins/signing.html

Terraform has been successfully initialized!

You may now begin working with Terraform. Try running "terraform plan" to see
any changes that are required for your infrastructure. All Terraform commands
should now work.

If you ever set or change modules or backend configuration for Terraform,
rerun this command to reinitialize your working directory. If you forget, other
commands will detect it and remind you to do so if necessary.
`

	got, err := filterTerraformInit(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should show provider versions
	if !strings.Contains(got, "hashicorp/aws v5.31.0") {
		t.Errorf("expected hashicorp/aws version, got:\n%s", got)
	}
	if !strings.Contains(got, "hashicorp/random v3.6.0") {
		t.Errorf("expected hashicorp/random version, got:\n%s", got)
	}
	if !strings.Contains(got, "hashicorp/null v3.2.2") {
		t.Errorf("expected hashicorp/null version, got:\n%s", got)
	}

	// Should show success
	if !strings.Contains(got, "Terraform has been successfully initialized!") {
		t.Errorf("expected success message, got:\n%s", got)
	}

	// Should NOT contain noise
	if strings.Contains(got, "Partner and community") {
		t.Errorf("should not contain partner signing info, got:\n%s", got)
	}
	if strings.Contains(got, "You may now begin") {
		t.Errorf("should not contain post-init instructions, got:\n%s", got)
	}
	if strings.Contains(got, "backend") {
		t.Errorf("should not contain backend config details, got:\n%s", got)
	}

	// Token savings >= 60%
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 60.0 {
		t.Errorf("expected >=60%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
}

func TestFilterTerraformInitWithError(t *testing.T) {
	raw := `
Initializing the backend...

Initializing provider plugins...
- Finding hashicorp/aws versions matching "~> 5.0"...

Error: Failed to query available provider packages

Could not retrieve the list of available versions for provider
hashicorp/aws: could not connect to registry.terraform.io:
connection refused.
`

	got, err := filterTerraformInit(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "Error: Failed to query available provider packages") {
		t.Errorf("expected error preserved, got:\n%s", got)
	}
}

func TestFilterTerraformInitEmpty(t *testing.T) {
	got, err := filterTerraformInit("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
