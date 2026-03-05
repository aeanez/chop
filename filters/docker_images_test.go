package filters

import (
	"strings"
	"testing"
)

func TestFilterDockerImages(t *testing.T) {
	raw := `REPOSITORY               TAG        IMAGE ID       CREATED         SIZE
nginx                    latest     a1b2c3d4e5f6   2 days ago      187MB
postgres                 15         b2c3d4e5f6a7   3 days ago      412MB
redis                    7-alpine   c3d4e5f6a7b8   5 days ago      30.2MB
node                     20-slim    d4e5f6a7b8c9   1 week ago      228MB
grafana/grafana          latest     e5f6a7b8c9d0   2 weeks ago     422MB
prom/prometheus          latest     f6a7b8c9d0e1   2 weeks ago     262MB
myapp                    dev        1234567890ab   3 hours ago     345MB
myapp                    latest     234567890abc   1 day ago       342MB
<none>                   <none>     34567890abcd   4 days ago      567MB
<none>                   <none>     4567890abcde   5 days ago      234MB
alpine                   3.18       567890abcdef   3 weeks ago     7.34MB
ubuntu                   22.04      67890abcdef0   1 month ago     77.8MB`

	got, err := filterDockerImages(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain tagged images
	if !strings.Contains(got, "nginx:latest") {
		t.Errorf("expected nginx:latest, got:\n%s", got)
	}
	if !strings.Contains(got, "redis:7-alpine") {
		t.Errorf("expected redis:7-alpine, got:\n%s", got)
	}

	// Should NOT contain <none> images (since tagged ones exist)
	if strings.Contains(got, "<none>") {
		t.Errorf("expected <none> images filtered out, got:\n%s", got)
	}

	// Should have total count (including none images)
	if !strings.Contains(got, "12 images total") {
		t.Errorf("expected '12 images total', got:\n%s", got)
	}

	// Each line should have size
	if !strings.Contains(got, "187MB") {
		t.Errorf("expected sizes in output, got:\n%s", got)
	}

	// Token savings >= 50%
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 50.0 {
		t.Errorf("expected >=50%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
}

func TestFilterDockerImagesOnlyNone(t *testing.T) {
	raw := `REPOSITORY   TAG       IMAGE ID       CREATED       SIZE
<none>       <none>    abc123def456   1 day ago     234MB
<none>       <none>    def456789012   2 days ago    567MB`

	got, err := filterDockerImages(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// When only <none> images exist, show them
	if !strings.Contains(got, "<none>") {
		t.Errorf("expected <none> images when no tagged images exist, got:\n%s", got)
	}

	if !strings.Contains(got, "2 images total") {
		t.Errorf("expected '2 images total', got:\n%s", got)
	}
}

func TestFilterDockerImagesEmpty(t *testing.T) {
	got, err := filterDockerImages("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty output, got %q", got)
	}
}
