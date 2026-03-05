package hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

var supportedCommands = map[string]bool{
	"git": true, "npm": true, "npx": true, "pnpm": true, "yarn": true, "bun": true,
	"docker": true, "dotnet": true, "kubectl": true, "helm": true, "terraform": true,
	"cargo": true, "go": true, "tsc": true, "eslint": true, "biome": true,
	"gh": true, "grep": true, "rg": true, "curl": true, "http": true,
	"aws": true, "az": true, "gcloud": true, "mvn": true, "gradle": true, "gradlew": true,
	"ng": true, "nx": true, "pytest": true, "pip": true, "pip3": true, "uv": true,
	"mypy": true, "ruff": true, "flake8": true, "pylint": true,
	"bundle": true, "bundler": true, "rspec": true, "rubocop": true,
	"composer": true, "make": true, "cmake": true,
	"gcc": true, "g++": true, "cc": true, "c++": true, "clang": true, "clang++": true,
	"ping": true, "ps": true, "ss": true, "netstat": true, "df": true, "du": true,
}

// shellBuiltins are commands that should never be wrapped.
var shellBuiltins = []string{
	"cd ", "export ", "source ", "echo ", "printf ", "set ", "unset ", "alias ", "eval ",
}

// compoundOperators are shell operators that indicate compound commands.
var compoundOperators = []string{"|", ">", ">>", "<", "&&", "||", ";"}

// hookInput represents the JSON payload received from Claude Code's PreToolUse hook.
type hookInput struct {
	SessionID     string          `json:"session_id"`
	Cwd           string          `json:"cwd"`
	HookEventName string          `json:"hook_event_name"`
	ToolName      string          `json:"tool_name"`
	ToolInput     json.RawMessage `json:"tool_input"`
}

type toolInput struct {
	Command string `json:"command"`
}

type hookOutput struct {
	HookSpecificOutput hookSpecificOutput `json:"hookSpecificOutput"`
}

type hookSpecificOutput struct {
	HookEventName      string    `json:"hookEventName"`
	PermissionDecision string    `json:"permissionDecision"`
	UpdatedInput       toolInput `json:"updatedInput"`
}

// RunHook reads a Claude Code PreToolUse hook payload from stdin,
// checks if the command should be wrapped with chop, and outputs
// modified JSON on stdout. Always exits 0.
func RunHook() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(0)
	}

	output, shouldModify := processHookInput(input)
	if shouldModify {
		fmt.Print(string(output))
	}
	// If not modifying, output nothing (passthrough)
}

// processHookInput parses the hook JSON and determines whether to wrap the command.
// Returns (outputJSON, true) if the command should be modified, or (nil, false) for passthrough.
func processHookInput(input []byte) ([]byte, bool) {
	var h hookInput
	if err := json.Unmarshal(input, &h); err != nil {
		return nil, false
	}

	if h.ToolName != "Bash" {
		return nil, false
	}

	var ti toolInput
	if err := json.Unmarshal(h.ToolInput, &ti); err != nil {
		return nil, false
	}

	command := strings.TrimSpace(ti.Command)
	if command == "" {
		return nil, false
	}

	// Already wrapped with chop
	if strings.HasPrefix(command, "chop ") {
		return nil, false
	}

	// Starts with a dot (source shorthand)
	if strings.HasPrefix(command, ". ") {
		return nil, false
	}

	// Shell builtins
	for _, prefix := range shellBuiltins {
		if strings.HasPrefix(command, prefix) {
			return nil, false
		}
	}

	// Compound commands — check for shell operators
	for _, op := range compoundOperators {
		if strings.Contains(command, op) {
			return nil, false
		}
	}

	// Extract base command name
	baseCmd := command
	if idx := strings.IndexByte(command, ' '); idx != -1 {
		baseCmd = command[:idx]
	}

	// Strip path prefix (e.g., /usr/bin/git -> git)
	if lastSlash := strings.LastIndexByte(baseCmd, '/'); lastSlash != -1 {
		baseCmd = baseCmd[lastSlash+1:]
	}

	if !supportedCommands[baseCmd] {
		return nil, false
	}

	out := hookOutput{
		HookSpecificOutput: hookSpecificOutput{
			HookEventName:      "PreToolUse",
			PermissionDecision: "allow",
			UpdatedInput: toolInput{
				Command: "chop " + command,
			},
		},
	}

	data, err := json.Marshal(out)
	if err != nil {
		return nil, false
	}

	return data, true
}
