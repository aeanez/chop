package discover

import (
	"bufio"
	"strings"
	"testing"
)

func TestScanReader_CountsUnwrappedCommands(t *testing.T) {
	lines := []string{
		// Unwrapped git — should count
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"git status"}}]}}`,
		// Unwrapped npm — should count
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"npm install"}}]}}`,
		// Another git — should count
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"git diff"}}]}}`,
		// Already wrapped — should NOT count
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"chop git log"}}]}}`,
		// Compound command — should NOT count
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"git status && git diff"}}]}}`,
		// Shell builtin — should NOT count
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"cd /tmp"}}]}}`,
		// Unsupported command — should NOT count
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"ls -la"}}]}}`,
		// Non-Bash tool — should NOT count
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Read","input":{"path":"/foo"}}]}}`,
		// Docker — should count
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"docker ps"}}]}}`,
		// Pipe — should NOT count
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"docker ps | grep foo"}}]}}`,
		// Malformed JSON — should be skipped silently
		`{"broken json`,
		// Empty line — should be skipped
		``,
	}

	input := strings.Join(lines, "\n")
	scanner := bufio.NewScanner(strings.NewReader(input))
	counts := make(map[string]int)
	ScanReader(scanner, counts)

	if counts["git"] != 2 {
		t.Errorf("expected git=2, got %d", counts["git"])
	}
	if counts["npm"] != 1 {
		t.Errorf("expected npm=1, got %d", counts["npm"])
	}
	if counts["docker"] != 1 {
		t.Errorf("expected docker=1, got %d", counts["docker"])
	}
	if len(counts) != 3 {
		t.Errorf("expected 3 unique commands, got %d: %v", len(counts), counts)
	}
}

func TestScanReader_SkipsWrappedCommands(t *testing.T) {
	lines := []string{
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"chop git status"}}]}}`,
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"chop npm test"}}]}}`,
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"chop docker ps"}}]}}`,
	}

	input := strings.Join(lines, "\n")
	scanner := bufio.NewScanner(strings.NewReader(input))
	counts := make(map[string]int)
	ScanReader(scanner, counts)

	if len(counts) != 0 {
		t.Errorf("expected 0 commands for all-wrapped input, got %d: %v", len(counts), counts)
	}
}

func TestScanReader_SkipsCompoundAndBuiltins(t *testing.T) {
	lines := []string{
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"echo hello"}}]}}`,
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"export FOO=bar"}}]}}`,
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"git log > out.txt"}}]}}`,
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"npm test || true"}}]}}`,
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"git status; git diff"}}]}}`,
	}

	input := strings.Join(lines, "\n")
	scanner := bufio.NewScanner(strings.NewReader(input))
	counts := make(map[string]int)
	ScanReader(scanner, counts)

	if len(counts) != 0 {
		t.Errorf("expected 0 commands, got %d: %v", len(counts), counts)
	}
}

func TestBuildResult_SortsByCount(t *testing.T) {
	counts := map[string]int{
		"npm":    3,
		"git":    10,
		"docker": 5,
	}

	result := buildResult(2, counts)

	if result.FilesScanned != 2 {
		t.Errorf("expected 2 files, got %d", result.FilesScanned)
	}
	if result.TotalCalls != 18 {
		t.Errorf("expected 18 total calls, got %d", result.TotalCalls)
	}
	if len(result.Counts) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(result.Counts))
	}
	if result.Counts[0].Name != "git" || result.Counts[0].Count != 10 {
		t.Errorf("expected git=10 first, got %s=%d", result.Counts[0].Name, result.Counts[0].Count)
	}
	if result.Counts[1].Name != "docker" || result.Counts[1].Count != 5 {
		t.Errorf("expected docker=5 second, got %s=%d", result.Counts[1].Name, result.Counts[1].Count)
	}
	if result.Counts[2].Name != "npm" || result.Counts[2].Count != 3 {
		t.Errorf("expected npm=3 third, got %s=%d", result.Counts[2].Name, result.Counts[2].Count)
	}
}

func TestExtractBashCommand_FlatFormat(t *testing.T) {
	line := `{"type":"tool_use","name":"Bash","input":{"command":"git status"}}`
	cmd := extractBashCommand(line)
	if cmd != "git status" {
		t.Errorf("expected 'git status', got %q", cmd)
	}
}

func TestExtractBashCommand_NestedFormat(t *testing.T) {
	line := `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"docker ps"}}]}}`
	cmd := extractBashCommand(line)
	if cmd != "docker ps" {
		t.Errorf("expected 'docker ps', got %q", cmd)
	}
}

func TestExtractBashCommand_NonBash(t *testing.T) {
	line := `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Read","input":{"path":"/foo"}}]}}`
	cmd := extractBashCommand(line)
	if cmd != "" {
		t.Errorf("expected empty, got %q", cmd)
	}
}

func TestExtractBashCommand_MalformedJSON(t *testing.T) {
	cmd := extractBashCommand(`{broken`)
	if cmd != "" {
		t.Errorf("expected empty for malformed JSON, got %q", cmd)
	}
}

func TestProcessLine_PathPrefixStripped(t *testing.T) {
	line := `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"/usr/bin/git status"}}]}}`
	counts := make(map[string]int)
	processLine(line, counts)
	if counts["git"] != 1 {
		t.Errorf("expected git=1 after path strip, got %d", counts["git"])
	}
}
