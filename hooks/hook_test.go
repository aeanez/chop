package hooks

import (
	"encoding/json"
	"testing"
)

func makeInput(command string) []byte {
	input := map[string]interface{}{
		"session_id":      "test-session",
		"cwd":             "/tmp",
		"hook_event_name": "PreToolUse",
		"tool_name":       "Bash",
		"tool_input": map[string]string{
			"command": command,
		},
	}
	data, _ := json.Marshal(input)
	return data
}

func TestSupportedCommandGetsPrepended(t *testing.T) {
	tests := []struct {
		cmd      string
		expected string
	}{
		{"npm test", "chop npm test"},
		{"git status", "chop git status"},
		{"docker ps", "chop docker ps"},
		{"kubectl get pods", "chop kubectl get pods"},
		{"cargo build", "chop cargo build"},
		{"go test ./...", "chop go test ./..."},
		{"curl https://api.io", "chop curl https://api.io"},
		{"dotnet build", "chop dotnet build"},
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			output, shouldModify := processHookInput(makeInput(tt.cmd))
			if !shouldModify {
				t.Fatalf("expected command to be modified: %s", tt.cmd)
			}

			var result hookOutput
			if err := json.Unmarshal(output, &result); err != nil {
				t.Fatalf("failed to parse output JSON: %v", err)
			}

			if result.HookSpecificOutput.UpdatedInput.Command != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result.HookSpecificOutput.UpdatedInput.Command)
			}
			if result.HookSpecificOutput.PermissionDecision != "allow" {
				t.Errorf("expected permission 'allow', got %q", result.HookSpecificOutput.PermissionDecision)
			}
			if result.HookSpecificOutput.HookEventName != "PreToolUse" {
				t.Errorf("expected hookEventName 'PreToolUse', got %q", result.HookSpecificOutput.HookEventName)
			}
		})
	}
}

func TestAlreadyChoppedPassthrough(t *testing.T) {
	_, shouldModify := processHookInput(makeInput("chop git status"))
	if shouldModify {
		t.Error("should not modify already-chopped command")
	}
}

func TestPipePassthrough(t *testing.T) {
	tests := []string{
		"git log | head -10",
		"docker ps | grep running",
		"cat file.txt | wc -l",
	}
	for _, cmd := range tests {
		t.Run(cmd, func(t *testing.T) {
			_, shouldModify := processHookInput(makeInput(cmd))
			if shouldModify {
				t.Errorf("should not modify pipe command: %s", cmd)
			}
		})
	}
}

func TestRedirectPassthrough(t *testing.T) {
	tests := []string{
		"git diff > output.txt",
		"echo hello >> log.txt",
		"docker run < input.txt",
	}
	for _, cmd := range tests {
		t.Run(cmd, func(t *testing.T) {
			_, shouldModify := processHookInput(makeInput(cmd))
			if shouldModify {
				t.Errorf("should not modify redirect command: %s", cmd)
			}
		})
	}
}

func TestCompoundCommandPassthrough(t *testing.T) {
	tests := []string{
		"git add . && git commit -m 'test'",
		"npm install || echo failed",
		"cd /tmp; ls",
	}
	for _, cmd := range tests {
		t.Run(cmd, func(t *testing.T) {
			_, shouldModify := processHookInput(makeInput(cmd))
			if shouldModify {
				t.Errorf("should not modify compound command: %s", cmd)
			}
		})
	}
}

func TestUnsupportedCommandPassthrough(t *testing.T) {
	tests := []string{
		"vim file.txt",
		"nano config.yml",
		"cat readme.md",
		"ls -la",
		"mkdir newdir",
		"rm -rf temp",
	}
	for _, cmd := range tests {
		t.Run(cmd, func(t *testing.T) {
			_, shouldModify := processHookInput(makeInput(cmd))
			if shouldModify {
				t.Errorf("should not modify unsupported command: %s", cmd)
			}
		})
	}
}

func TestEmptyCommandPassthrough(t *testing.T) {
	_, shouldModify := processHookInput(makeInput(""))
	if shouldModify {
		t.Error("should not modify empty command")
	}
}

func TestShellBuiltinPassthrough(t *testing.T) {
	tests := []string{
		"cd /tmp",
		"export FOO=bar",
		"source ~/.bashrc",
		". ~/.bashrc",
		"echo hello world",
		"printf '%s\\n' hello",
		"set -e",
		"unset FOO",
		"alias ll='ls -la'",
		"eval some_command",
	}
	for _, cmd := range tests {
		t.Run(cmd, func(t *testing.T) {
			_, shouldModify := processHookInput(makeInput(cmd))
			if shouldModify {
				t.Errorf("should not modify shell builtin: %s", cmd)
			}
		})
	}
}

func TestNonBashToolPassthrough(t *testing.T) {
	input := map[string]interface{}{
		"session_id":      "test-session",
		"cwd":             "/tmp",
		"hook_event_name": "PreToolUse",
		"tool_name":       "Read",
		"tool_input": map[string]string{
			"file_path": "/some/file.txt",
		},
	}
	data, _ := json.Marshal(input)
	_, shouldModify := processHookInput(data)
	if shouldModify {
		t.Error("should not modify non-Bash tool")
	}
}

func TestInvalidJSONPassthrough(t *testing.T) {
	_, shouldModify := processHookInput([]byte("not json"))
	if shouldModify {
		t.Error("should not modify invalid JSON")
	}
}
