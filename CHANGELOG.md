# Changelog

All notable changes to chop are documented here.

---

## [v1.6.0] - 2026-03-12

### Added
- **User-defined custom filters** - define output compression rules for any command without writing Go code
  - `keep: [regex...]` - keep only lines matching at least one pattern
  - `drop: [regex...]` - remove lines matching any pattern
  - `head: N` / `tail: N` - truncate to first/last N lines
  - `exec: script` - pipe output through any shell command or script (e.g. `jq .`, `python3 filter.py`)
- **Two-level filter config** - global `~/.config/chop/filters.yml` + local `.chop-filters.yml` per project; local always wins
- **`chop filter add`** - add or update a filter from the CLI, no YAML editing required
  - `--keep`, `--drop`, `--head`, `--tail`, `--exec` flags
  - `--local` flag to write to the project-level file
  - Pass raw regex patterns - no manual escaping needed (`\s` not `\\s`)
- **`chop filter remove`** - remove a filter by command name (`--local` for project-level)
- **`chop filter init --local`** - create a starter `.chop-filters.yml` in the current directory
- **`chop config init`** - create a starter global `~/.config/chop/config.yml`
- **`chop filter test`** - test a custom filter against stdin without running the actual command

### Fixed
- `exec` filters with system commands (`jq`, `python3`, etc.) or scripts with arguments now work correctly - the previous `os.Stat` check incorrectly rejected anything that wasn't a plain file path
- Invalid regex patterns now warn to stderr instead of silently being ignored

### Changed
- Replaced all em-dashes with regular dashes across the codebase

---

## [v1.5.0] - 2026-03-11

### Added
- Expanded filter coverage across more commands and subcommands
- Unchopped report UX improvements: cleaner output, better section headings
- `--skip` / `--unskip` flags to mark commands as intentionally unfiltered
- `--delete` flag to permanently remove tracking records for a command
- `--verbose` flag for untruncated command names in unchopped report

---

## [v1.4.0] - 2026-03-11

### Added
- **`chop gain --unchopped`** - identify commands that pass through without compression, showing candidates for new filters
- Two report sections: commands with no filter registered vs. filters that never triggered

---

## [v1.3.0] - 2026-03-11

### Added
- **Log pattern compression** - groups structurally similar log lines by replacing variable parts (UUIDs, IPs, timestamps, `key=value` pairs) with a fingerprint, then shows a representative line with a repeat count
- Errors and warnings are always shown in full and floated to the top
- Falls back to exact-match deduplication when no repeating patterns are found
- Applies to `cat`, `tail`, `less`, `more`, and any log-producing command

---

## [v1.2.2] - 2026-03-10

### Fixed
- Panic on `docker ps` with a custom `--format table` flag

---

## [v1.2.1] - 2026-03-09

### Fixed
- Gain stats now use local timezone instead of UTC

---

## [v1.2.0] - 2026-03-09

### Added
- Filter support for `npx playwright`, `npx tsc`, `npx ng`, `acli jira`, `node`, `ls`, and `find`

---

## [v1.1.0] - 2026-03-09

### Added
- **Subcommand-level disabled config** - disable `"git diff"` without affecting `git status`
- **Local `.chop.yml`** - per-project overrides managed via `chop local add/remove/clear`; local list replaces the global one when present
- Section-aware git status filter

---

## [v1.0.5] - 2026-03-09

### Fixed
- Gain stats now use calendar-based periods (week = Mon-Sun, month starts on 1st, year starts Jan 1)

---

## [v1.0.4] - 2026-03-09

### Added
- **`chop doctor`** - detects and fixes common issues such as hook path mismatches after moves or updates

---

## [v1.0.3] - 2026-03-09

### Added
- Weekly, monthly, and yearly metrics to `chop gain`
- Fixed Windows install path detection

---

## [v1.0.2] - 2026-03-09

### Fixed
- Auto-add to PATH on Windows during install

---

## [v1.0.1] - 2026-03-07

### Fixed
- Handle git global flags (`-C`, `--no-pager`, `-c key=val`, etc.) before subcommand matching so `git -C /path status` is correctly routed

---

## [v1.0.0] - 2026-03-07

Initial stable release.

### Highlights
- 60+ built-in filters covering Git, npm/pnpm/yarn/bun, Angular/Nx, .NET, Rust, Go, Python, Java, Ruby, PHP, Docker, Kubernetes, Terraform, cloud CLIs, HTTP, and more
- Auto-detection for JSON, CSV, tables, and log output on unknown commands
- Token tracking with `chop gain` - per-command savings stored in a local SQLite database
- Claude Code hook integration (`chop init --global`) - wraps every Bash command automatically
- Windows support with PowerShell installer and native PATH management
- `chop update` for self-updating
- `chop uninstall` / `chop reset` for clean removal
