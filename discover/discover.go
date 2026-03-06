package discover

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// supportedCommands mirrors hooks.supportedCommands to avoid circular deps.
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

// shellBuiltins are command prefixes that should never be wrapped.
var shellBuiltins = []string{
	"cd ", "export ", "source ", "echo ", "printf ", "set ", "unset ", "alias ", "eval ",
}

// compoundOperators indicate compound commands that should be skipped.
var compoundOperators = []string{"|", ">", ">>", "<", "&&", "||", ";"}

// commandCount tracks how many unwrapped calls a command had.
type commandCount struct {
	Name  string
	Count int
}

// Result holds the scan results for testing.
type Result struct {
	FilesScanned int
	Counts       []commandCount
	TotalCalls   int
}

// Run scans Claude Code session logs and prints missed savings opportunities.
func Run() {
	result := Scan(defaultProjectsDir())
	PrintReport(result)
}

// defaultProjectsDir returns the ~/.claude/projects/ path.
func defaultProjectsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "projects")
}

// Scan finds JSONL files and analyzes them for missed chop opportunities.
func Scan(projectsDir string) Result {
	if projectsDir == "" {
		return Result{}
	}

	files := findJSONLFiles(projectsDir)
	if len(files) == 0 {
		return Result{}
	}

	fmt.Fprintf(os.Stderr, "scanning %d files...\n", len(files))

	counts := make(map[string]int)
	for _, f := range files {
		scanFile(f, counts)
	}

	return buildResult(len(files), counts)
}

// ScanReader scans a single reader for missed chop opportunities (for testing).
func ScanReader(r *bufio.Scanner, counts map[string]int) {
	for r.Scan() {
		line := r.Text()
		processLine(line, counts)
	}
}

// findJSONLFiles returns .jsonl files modified within the last 30 days.
func findJSONLFiles(root string) []string {
	cutoff := time.Now().AddDate(0, 0, -30)
	var files []string

	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip inaccessible dirs
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".jsonl") {
			return nil
		}
		if info.ModTime().Before(cutoff) {
			return nil
		}
		files = append(files, path)
		return nil
	})

	return files
}

// scanFile reads a JSONL file line-by-line and updates counts.
func scanFile(path string, counts map[string]int) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	// Allow larger lines — session logs can have big entries
	scanner.Buffer(make([]byte, 0, 256*1024), 2*1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		processLine(line, counts)
	}
}

// processLine checks a single JSONL line for unwrapped Bash commands.
func processLine(line string, counts map[string]int) {
	// Quick string check before expensive JSON parsing
	if !strings.Contains(line, "Bash") {
		return
	}

	cmd := extractBashCommand(line)
	if cmd == "" {
		return
	}

	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return
	}

	// Already wrapped
	if strings.HasPrefix(cmd, "chop ") {
		return
	}

	// Shell builtins
	for _, prefix := range shellBuiltins {
		if strings.HasPrefix(cmd, prefix) {
			return
		}
	}

	// Compound commands
	for _, op := range compoundOperators {
		if strings.Contains(cmd, op) {
			return
		}
	}

	// Extract base command
	baseCmd := cmd
	if idx := strings.IndexByte(cmd, ' '); idx != -1 {
		baseCmd = cmd[:idx]
	}

	// Strip path prefix
	if lastSlash := strings.LastIndexByte(baseCmd, '/'); lastSlash != -1 {
		baseCmd = baseCmd[lastSlash+1:]
	}

	if supportedCommands[baseCmd] {
		counts[baseCmd]++
	}
}

// extractBashCommand tries to extract the command string from a JSONL line.
// The format varies but we look for tool_use entries with name "Bash".
func extractBashCommand(line string) string {
	// Try the nested message.content format first
	var nested struct {
		Message *struct {
			Content []struct {
				Type  string `json:"type"`
				Name  string `json:"name"`
				Input struct {
					Command string `json:"command"`
				} `json:"input"`
			} `json:"content"`
		} `json:"message"`
		// Also try flat format
		Type  string `json:"type"`
		Name  string `json:"name"`
		Input *struct {
			Command string `json:"command"`
		} `json:"input"`
	}

	if err := json.Unmarshal([]byte(line), &nested); err != nil {
		return ""
	}

	// Check nested message.content[] format
	if nested.Message != nil {
		for _, c := range nested.Message.Content {
			if c.Type == "tool_use" && c.Name == "Bash" && c.Input.Command != "" {
				return c.Input.Command
			}
		}
	}

	// Check flat format
	if nested.Name == "Bash" && nested.Input != nil && nested.Input.Command != "" {
		return nested.Input.Command
	}

	return ""
}

// buildResult converts the counts map into a sorted Result.
func buildResult(filesScanned int, counts map[string]int) Result {
	var sorted []commandCount
	total := 0
	for name, count := range counts {
		sorted = append(sorted, commandCount{Name: name, Count: count})
		total += count
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Count > sorted[j].Count
	})
	return Result{
		FilesScanned: filesScanned,
		Counts:       sorted,
		TotalCalls:   total,
	}
}

// PrintReport outputs the discover summary.
func PrintReport(r Result) {
	fmt.Println("chop discover — missed savings opportunities")
	fmt.Println()

	if r.FilesScanned == 0 {
		fmt.Println("No session files found.")
		return
	}

	fmt.Printf("Scanned %d session files\n", r.FilesScanned)
	fmt.Println()

	if len(r.Counts) == 0 {
		fmt.Println("All commands are already wrapped with chop. Nice!")
		return
	}

	fmt.Printf("  %-14s %5s   %s\n", "COMMAND", "CALLS", "STATUS")
	for _, c := range r.Counts {
		fmt.Printf("  %-14s %5d   not wrapped\n", c.Name, c.Count)
	}
	fmt.Println()
	fmt.Printf("%d commands could benefit from chop (%d total calls)\n", len(r.Counts), r.TotalCalls)
	fmt.Println()
	fmt.Println("Run 'chop init --global' to auto-wrap all commands via Claude Code hook.")
}
