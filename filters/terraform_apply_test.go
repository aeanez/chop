package filters

import (
	"strings"
	"testing"
)

func TestFilterTerraformApply(t *testing.T) {
	raw := `
aws_security_group.web: Modifying... [id=sg-12345678]
aws_security_group.web: Modifications complete after 3s [id=sg-12345678]
aws_instance.web: Creating...
aws_instance.web: Still creating... [10s elapsed]
aws_instance.web: Still creating... [20s elapsed]
aws_instance.web: Still creating... [30s elapsed]
aws_instance.web: Still creating... [40s elapsed]
aws_instance.web: Still creating... [50s elapsed]
aws_instance.web: Still creating... [1m0s elapsed]
aws_instance.web: Still creating... [1m10s elapsed]
aws_instance.web: Creation complete after 1m15s [id=i-0abcdef1234567890]
aws_s3_bucket.assets: Creating...
aws_s3_bucket.assets: Creation complete after 2s [id=my-assets-bucket]
aws_db_instance.main: Modifying... [id=mydb]
aws_db_instance.main: Still modifying... [id=mydb, 10s elapsed]
aws_db_instance.main: Still modifying... [id=mydb, 20s elapsed]
aws_db_instance.main: Still modifying... [id=mydb, 30s elapsed]
aws_db_instance.main: Still modifying... [id=mydb, 40s elapsed]
aws_db_instance.main: Still modifying... [id=mydb, 50s elapsed]
aws_db_instance.main: Still modifying... [id=mydb, 1m0s elapsed]
aws_db_instance.main: Still modifying... [id=mydb, 1m10s elapsed]
aws_db_instance.main: Still modifying... [id=mydb, 1m20s elapsed]
aws_db_instance.main: Still modifying... [id=mydb, 1m30s elapsed]
aws_db_instance.main: Still modifying... [id=mydb, 1m40s elapsed]
aws_db_instance.main: Still modifying... [id=mydb, 1m50s elapsed]
aws_db_instance.main: Still modifying... [id=mydb, 2m0s elapsed]
aws_db_instance.main: Modifications complete after 2m5s [id=mydb]
aws_route53_record.www: Destroying... [id=Z1234567890_www.example.com_A]
aws_route53_record.www: Still destroying... [id=Z1234567890_www.example.com_A, 10s elapsed]
aws_route53_record.www: Still destroying... [id=Z1234567890_www.example.com_A, 20s elapsed]
aws_route53_record.www: Destruction complete after 25s

Apply complete! Resources: 2 added, 2 changed, 1 destroyed.

Outputs:

instance_id = "i-0abcdef1234567890"
bucket_name = "my-assets-bucket"
`

	got, err := filterTerraformApply(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have resource lines with timing
	if !strings.Contains(got, "aws_instance.web: created (1m15s)") {
		t.Errorf("expected aws_instance.web created with timing, got:\n%s", got)
	}
	if !strings.Contains(got, "aws_s3_bucket.assets: created (2s)") {
		t.Errorf("expected aws_s3_bucket.assets created with timing, got:\n%s", got)
	}
	if !strings.Contains(got, "aws_security_group.web: updated (3s)") {
		t.Errorf("expected aws_security_group.web updated with timing, got:\n%s", got)
	}
	if !strings.Contains(got, "aws_db_instance.main: updated (2m5s)") {
		t.Errorf("expected aws_db_instance.main updated with timing, got:\n%s", got)
	}
	if !strings.Contains(got, "aws_route53_record.www: destroyed (25s)") {
		t.Errorf("expected aws_route53_record.www destroyed with timing, got:\n%s", got)
	}

	// Should have summary
	if !strings.Contains(got, "Apply complete! Resources: 2 added, 2 changed, 1 destroyed") {
		t.Errorf("expected apply summary, got:\n%s", got)
	}

	// Should NOT contain "Still creating" lines
	if strings.Contains(got, "Still") {
		t.Errorf("should not contain 'Still' progress lines, got:\n%s", got)
	}

	// Token savings >= 70%
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 70.0 {
		t.Errorf("expected >=70%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
}

func TestFilterTerraformApplyWithErrors(t *testing.T) {
	raw := `
aws_instance.web: Creating...
aws_instance.web: Still creating... [10s elapsed]

Error: Error launching source instance: VPCIdNotSpecified: No default VPC for this user

  on main.tf line 10, in resource "aws_instance" "web":
  10: resource "aws_instance" "web" {
`

	got, err := filterTerraformApply(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "Error:") {
		t.Errorf("expected error preserved, got:\n%s", got)
	}
}

func TestFilterTerraformApplyEmpty(t *testing.T) {
	got, err := filterTerraformApply("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
