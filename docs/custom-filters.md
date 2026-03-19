# Custom Filters

Define your own output compression rules for any command — no Go code required. Filters live in `~/.config/chop/filters.yml` (global) or `.chop-filters.yml` in your project directory (local).

## Quick start

```bash
chop filter new "myctl deploy"   # scaffold a filter + guided workflow
```

Or create the config file first:

```bash
chop filter init                 # global ~/.config/chop/filters.yml
chop filter init --local         # local .chop-filters.yml in cwd
```

## Managing filters

| Command | Description |
|---------|-------------|
| `chop filter` | List all active custom filters |
| `chop filter path` | Show the global filters file path |
| `chop filter init` | Create starter global filters.yml |
| `chop filter init --local` | Create starter .chop-filters.yml in cwd |
| `chop filter new <cmd>` | Scaffold a filter + guided workflow |
| `chop filter add <cmd> [flags]` | Add or update a filter |
| `chop filter remove <cmd>` | Remove a filter |
| `chop filter remove <cmd> --local` | Remove a project-level filter |
| `chop filter test <cmd>` | Test a filter against stdin |

Local filters are merged on top of global ones — **local always wins on conflict**.

## Guided filter creation

`chop filter new` is the recommended starting point. It scaffolds a commented-out entry in `filters.yml` and prints the four-step workflow:

```bash
chop filter new "myctl deploy"
# scaffolded filter for "myctl deploy" in ~/.config/chop/filters.yml
#
# next steps:
#   1. chop capture myctl deploy          — capture real output as a fixture
#   2. edit filters.yml                   — uncomment and tune the rules
#   3. chop diff myctl deploy             — preview compression before enabling
#   4. chop filter test myctl deploy      — verify against stdin
```

## Adding filters from the CLI

Use `chop filter add` with one or more rule flags:

```bash
chop filter add "myctl deploy" --keep "ERROR,WARN,deployed,^=" --drop "DEBUG,^\s*$"
chop filter add "ansible-playbook" --keep "^PLAY,^TASK,fatal,changed,^\s+ok=" --tail 20
chop filter add "custom-tool" --exec "~/.config/chop/scripts/custom-tool.sh"
chop filter add "make build" --keep "error:,warning:,^make\[" --tail 10 --local
```

| Flag | Description |
|------|-------------|
| `--keep "p1,p2"` | Comma-separated regex patterns — only keep matching lines |
| `--drop "p1,p2"` | Comma-separated regex patterns — remove matching lines |
| `--head N` | Keep first N lines (after drop/keep) |
| `--tail N` | Keep last N lines (after drop/keep) |
| `--exec script` | Pipe output through an external script or command |
| `--local` | Write to `.chop-filters.yml` in the current directory |

> **No manual escaping needed:** pass regex patterns as-is. chop handles escaping when writing the YAML. Use `\s` for whitespace, `\d` for digits, etc.
>
> ```bash
> # Correct
> chop filter add "mytool" --drop "^\s*$"
>
> # Wrong — double-escaping produces the wrong regex
> chop filter add "mytool" --drop "^\\s*$"
> ```

## Rules

Rules are applied in this order:

| Rule | Description |
|------|-------------|
| `drop` | Remove lines matching **any** pattern (applied first) |
| `keep` | Keep only lines matching **at least one** pattern |
| `head: N` | Keep first N lines (after drop/keep) |
| `tail: N` | Keep last N lines (after drop/keep) |
| `exec` | Pipe raw output through an external script (stdin → stdout) |

If both `head` and `tail` are set and output exceeds `head + tail` lines, a `... (N lines hidden)` separator is shown between them.

`exec` takes priority — when set, all other rules are ignored and the script receives the raw output on stdin. Supports any command in your shell (`jq .`, `python3 filter.py`, etc.).

## Editing the YAML directly

You can edit `filters.yml` directly. Note that backslashes **must** be escaped in YAML double-quoted strings (`\\s` in the file = `\s` in the regex). `chop filter add` handles this automatically.

```yaml
filters:
  # Keep only error/warning lines from a custom CLI tool
  "myctl deploy":
    keep: ["ERROR", "WARN", "deployed", "^="]
    drop: ["DEBUG", "^\\s*$"]      # \\s in YAML = \s in the regex

  # Show key task headers + last 20 lines of ansible output
  "ansible-playbook":
    keep: ["^PLAY", "^TASK", "fatal", "changed", "^\\s+ok="]
    tail: 20

  # Pipe output through any shell command
  "custom-tool":
    exec: "jq ."
```

## Testing filters

Test a filter against sample input without running the actual command:

```bash
# Linux/macOS
echo -e "DEBUG init\nINFO started\nERROR failed" | chop filter test myctl deploy

# Windows (PowerShell)
"DEBUG init`nINFO started`nERROR failed" | chop filter test myctl deploy
```

## Security

The `exec` field executes an arbitrary external command. For this reason:

- Filters in `~/.config/chop/filters.yml` are **trusted** — `exec` works.
- Filters in `.chop-filters.yml` (local, per-project) are **untrusted** — `exec` is silently skipped with a warning. This prevents a malicious repository from running arbitrary code on your machine.

To use `exec` for a project-specific command, define it in your global `filters.yml`.
