# Token Tracking

Every command run through chop is recorded in a local SQLite database. Use `chop gain` to query your savings.

## Overview

```bash
chop gain              # overall savings (today / week / month / year / total)
chop gain --since 7d   # stats for a specific time window
```

Time window formats: `30m`, `24h`, `7d`, `2w`, `1mo`.

```
$ chop gain
chop - token savings report

  today: 42 commands, 12,847 tokens saved
  week:  187 commands, 52,340 tokens saved
  month: 318 commands, 89,234 tokens saved
  year:  1,203 commands, 456,789 tokens saved
  total: 1,203 commands, 456,789 tokens saved (73.2% avg)
```

## History

```bash
chop gain --history                       # last 20 commands with per-command savings
chop gain --history --limit 100           # last 100 commands
chop gain --history --all                 # all recorded commands
chop gain --history --since 7d            # history filtered to last 7 days
chop gain --history --since 7d --all      # all commands in the last 7 days
chop gain --history --verbose             # full command strings + project group headers
chop gain --history --project <path>      # history for a specific project root
```

## Summaries

```bash
chop gain --summary    # per-command savings breakdown
chop gain --projects   # per-project savings breakdown (grouped by git root)
```

## The `!` marker

`chop gain --history` marks commands with `!` when they produced 0% savings. This is expected in two cases:

- **Write commands** (`git commit`, `git push`, `git add`, etc.) — near-zero output by design, nothing to compress.
- **Already-minimal output** — a `git log --oneline -5` or a `find` that returned one result is already compact.

To remove these entries and stop tracking them:

```bash
chop gain --no-track "git push"
chop gain --no-track "git commit"
```

## Suppressing tracking

```bash
chop gain --no-track "git push"      # delete records for X and stop tracking it
chop gain --resume-track "git push"  # re-enable tracking
chop gain --delete "git push"        # delete all records for X (tracking continues)
```

## Unchopped report

Identify commands that pass through without compression — candidates for new filters:

```bash
chop gain --unchopped            # commands with no filter coverage
chop gain --unchopped --verbose  # untruncated command names + full detail
chop gain --unchopped --skip X   # mark X as intentionally unfiltered (hides it)
chop gain --unchopped --unskip X # restore X to the candidates list
```

The report has two sections:

- **no filter registered** — output passes through raw; worth writing a filter if AVG tokens is high
- **filter registered, 0% runs** — filter exists but output was already minimal; no action needed

## Export

```bash
chop gain --export json           # export full history as JSON to stdout
chop gain --export csv            # export full history as CSV to stdout
chop gain --export json --since 7d  # scoped to a time window
```

Redirect to a file:

```bash
chop gain --export json > chop-history.json
chop gain --export csv  > chop-history.csv
```
