package filters

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestCompressJSONLargeObject(t *testing.T) {
	// Build a large API response with nested arrays and verbose values
	// Simulates a realistic user profile endpoint with embedded collections
	teammates := make([]interface{}, 20)
	for i := 0; i < 20; i++ {
		teammates[i] = map[string]interface{}{
			"id":    float64(100 + i),
			"name":  fmt.Sprintf("Team Member %d", i+1),
			"email": fmt.Sprintf("member%d@example.com", i+1),
			"role":  "contributor",
		}
	}
	auditLog := make([]interface{}, 30)
	for i := 0; i < 30; i++ {
		auditLog[i] = map[string]interface{}{
			"action":    fmt.Sprintf("action_%d", i),
			"timestamp": "2024-06-20T14:22:00Z",
			"ip":        "192.168.1.1",
			"details":   fmt.Sprintf("Performed action %d on resource xyz-abc-123", i),
		}
	}
	obj := map[string]interface{}{
		"id":          12345,
		"name":        "Test User",
		"email":       "test@example.com",
		"active":      true,
		"role":        "admin",
		"department":  "engineering",
		"location":    "New York",
		"timezone":    "America/New_York",
		"created_at":  "2024-01-15T10:30:00Z",
		"updated_at":  "2024-06-20T14:22:00Z",
		"last_login":  "2024-06-20T08:00:00Z",
		"login_count": 342,
		"preferences": map[string]interface{}{
			"theme":         "dark",
			"language":      "en",
			"notifications": true,
			"font_size":     14,
			"sidebar":       "collapsed",
		},
		"permissions": []interface{}{"read", "write", "admin", "deploy", "audit", "manage_users", "billing", "settings", "reports", "analytics"},
		"teams":       []interface{}{"backend", "platform", "sre", "security", "devops"},
		"teammates":   teammates,
		"audit_log":   auditLog,
		"metadata": map[string]interface{}{
			"source":     "ldap",
			"verified":   true,
			"risk_score": 0.12,
			"tags":       []interface{}{"senior", "lead", "oncall"},
			"nested_deep": map[string]interface{}{
				"level3_key": "should be summarized",
			},
		},
		"avatar_url":    "https://cdn.example.com/avatars/12345/profile-picture-large-high-resolution.png",
		"bio":           "This is a really long bio that exceeds fifty characters so it should be truncated by the compression algorithm to save tokens",
		"phone":         "+1-555-0123",
		"address":       "123 Main Street, Apt 4B, New York, NY 10001, United States of America",
		"api_key":       "sk-1234567890abcdef1234567890abcdef1234567890abcdef52",
		"session_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ",
	}

	raw, _ := json.MarshalIndent(obj, "", "  ")
	rawStr := string(raw)

	got, err := compressJSON(rawStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain key names
	if !strings.Contains(got, "id") {
		t.Error("expected output to contain key 'id'")
	}
	if !strings.Contains(got, "permissions") {
		t.Error("expected output to contain key 'permissions'")
	}
	// Should use type names for values
	if !strings.Contains(got, "number") {
		t.Error("expected type 'number' in output")
	}
	if !strings.Contains(got, "string") {
		t.Error("expected type 'string' in output")
	}

	// Token savings >= 70%
	rawTokens := len(strings.Fields(rawStr))
	gotTokens := len(strings.Fields(got))
	savings := 100.0 - float64(gotTokens)/float64(rawTokens)*100.0
	if savings < 70.0 {
		t.Errorf("expected >=70%% token savings, got %.1f%% (raw=%d, filtered=%d)\noutput:\n%s", savings, rawTokens, gotTokens, got)
	}
}

func TestCompressJSONLargeArray(t *testing.T) {
	// Build array of 50+ items
	items := make([]interface{}, 50)
	for i := 0; i < 50; i++ {
		items[i] = map[string]interface{}{
			"id":    float64(i + 1),
			"name":  fmt.Sprintf("User %d", i+1),
			"email": fmt.Sprintf("user%d@example.com", i+1),
			"age":   float64(20 + i),
		}
	}

	raw, _ := json.MarshalIndent(items, "", "  ")
	rawStr := string(raw)

	got, err := compressJSON(rawStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should show first element structure + count
	if !strings.Contains(got, "x50") {
		t.Errorf("expected output to contain 'x50', got:\n%s", got)
	}

	// Token savings >= 70%
	rawTokens := len(strings.Fields(rawStr))
	gotTokens := len(strings.Fields(got))
	savings := 100.0 - float64(gotTokens)/float64(rawTokens)*100.0
	if savings < 70.0 {
		t.Errorf("expected >=70%% token savings, got %.1f%% (raw=%d, filtered=%d)\noutput:\n%s", savings, rawTokens, gotTokens, got)
	}
}

func TestCompressJSONSmallSimple(t *testing.T) {
	raw := `{"status": "ok", "count": 3, "active": true}`

	got, err := compressJSON(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Small JSON should preserve actual values
	if !strings.Contains(got, "ok") {
		t.Errorf("expected small JSON to preserve value 'ok', got:\n%s", got)
	}
	if !strings.Contains(got, "3") {
		t.Errorf("expected small JSON to preserve value '3', got:\n%s", got)
	}
	if !strings.Contains(got, "true") {
		t.Errorf("expected small JSON to preserve value 'true', got:\n%s", got)
	}
}

func TestCompressJSONEmpty(t *testing.T) {
	got, err := compressJSON("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestCompressJSONInvalid(t *testing.T) {
	_, err := compressJSON("not json at all")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestCompressJSONNestedObject(t *testing.T) {
	raw := `{
		"data": {
			"users": [
				{"id": 1, "name": "Alice", "profile": {"bio": "hello", "avatar": "url"}},
				{"id": 2, "name": "Bob", "profile": {"bio": "world", "avatar": "url2"}},
				{"id": 3, "name": "Carol", "profile": {"bio": "test", "avatar": "url3"}},
				{"id": 4, "name": "Dave", "profile": {"bio": "demo", "avatar": "url4"}},
				{"id": 5, "name": "Eve", "profile": {"bio": "sample", "avatar": "url5"}},
				{"id": 6, "name": "Frank", "profile": {"bio": "example", "avatar": "url6"}}
			],
			"total": 100,
			"page": 1,
			"per_page": 6,
			"has_more": true,
			"query": "active"
		},
		"meta": {
			"request_id": "abc-123",
			"duration_ms": 42,
			"cache": "hit"
		}
	}`

	got, err := compressJSON(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should show array count
	if !strings.Contains(got, "x6") {
		t.Errorf("expected output to contain 'x6' for users array, got:\n%s", got)
	}
}

func TestCompressJSONLongStrings(t *testing.T) {
	longStr := strings.Repeat("a", 100)
	obj := map[string]interface{}{
		"short":   "hello",
		"long":    longStr,
		"another": "world",
		"extra1":  "padding1",
		"extra2":  "padding2",
		"extra3":  "padding3",
	}
	raw, _ := json.MarshalIndent(obj, "", "  ")

	got, err := compressJSON(string(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Long string should not appear in full
	if strings.Contains(got, longStr) {
		t.Error("expected long string to be absent or truncated")
	}
}
