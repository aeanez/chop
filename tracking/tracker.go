package tracking

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// Stats holds aggregate token savings statistics.
type Stats struct {
	TotalCommands    int
	TotalRawTokens   int
	TotalSavedTokens int
	OverallSavingsPct float64
	TodayCommands    int
	TodaySavedTokens int
}

// Record holds a single tracking entry.
type Record struct {
	Timestamp      string
	Command        string
	RawTokens      int
	FilteredTokens int
	SavingsPct     float64
}

var (
	db     *sql.DB
	dbOnce sync.Once
	dbErr  error
)

func dbPath() string {
	if p := os.Getenv("CHOP_DB_PATH"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".local", "share", "chop", "tracking.db")
}

// Init opens (or creates) the tracking database and ensures the schema exists.
func Init() error {
	dbOnce.Do(func() {
		path := dbPath()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			dbErr = err
			return
		}
		db, dbErr = sql.Open("sqlite", path)
		if dbErr != nil {
			return
		}
		_, dbErr = db.Exec(`CREATE TABLE IF NOT EXISTS tracking (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp TEXT NOT NULL,
			command TEXT NOT NULL,
			raw_tokens INTEGER NOT NULL,
			filtered_tokens INTEGER NOT NULL,
			savings_pct REAL NOT NULL
		)`)
	})
	return dbErr
}

// initForTest resets the singleton so tests can re-init with a new DB path.
func initForTest() {
	dbOnce = sync.Once{}
	db = nil
	dbErr = nil
}

// Track records a command's token savings. Silent on error.
func Track(command string, rawTokens, filteredTokens int) error {
	if err := Init(); err != nil {
		return err
	}
	var savingsPct float64
	if rawTokens > 0 {
		savingsPct = 100.0 - (float64(filteredTokens) / float64(rawTokens) * 100.0)
	}
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	_, err := db.Exec(
		`INSERT INTO tracking (timestamp, command, raw_tokens, filtered_tokens, savings_pct)
		 VALUES (?, ?, ?, ?, ?)`,
		now, command, rawTokens, filteredTokens, savingsPct,
	)
	return err
}

// GetStats returns aggregate statistics.
func GetStats() (Stats, error) {
	if err := Init(); err != nil {
		return Stats{}, err
	}
	var s Stats

	row := db.QueryRow(`SELECT COUNT(*), COALESCE(SUM(raw_tokens),0), COALESCE(SUM(raw_tokens - filtered_tokens),0) FROM tracking`)
	if err := row.Scan(&s.TotalCommands, &s.TotalRawTokens, &s.TotalSavedTokens); err != nil {
		return Stats{}, err
	}
	if s.TotalRawTokens > 0 {
		s.OverallSavingsPct = float64(s.TotalSavedTokens) / float64(s.TotalRawTokens) * 100.0
	}

	today := time.Now().UTC().Format("2006-01-02")
	row = db.QueryRow(
		`SELECT COUNT(*), COALESCE(SUM(raw_tokens - filtered_tokens),0) FROM tracking WHERE timestamp LIKE ?`,
		today+"%",
	)
	if err := row.Scan(&s.TodayCommands, &s.TodaySavedTokens); err != nil {
		return Stats{}, err
	}

	return s, nil
}

// GetHistory returns the last N tracking records in reverse chronological order.
func GetHistory(limit int) ([]Record, error) {
	if err := Init(); err != nil {
		return nil, err
	}
	rows, err := db.Query(
		`SELECT timestamp, command, raw_tokens, filtered_tokens, savings_pct
		 FROM tracking ORDER BY id DESC LIMIT ?`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.Timestamp, &r.Command, &r.RawTokens, &r.FilteredTokens, &r.SavingsPct); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

// Cleanup removes records older than the given number of days.
func Cleanup(days int) error {
	if err := Init(); err != nil {
		return err
	}
	cutoff := time.Now().UTC().AddDate(0, 0, -days).Format("2006-01-02 15:04:05")
	_, err := db.Exec(`DELETE FROM tracking WHERE timestamp < ?`, cutoff)
	return err
}

// CountTokens returns the word count of a string (whitespace-split).
func CountTokens(s string) int {
	return len(strings.Fields(s))
}

// FormatGain prints the gain summary report.
func FormatGain(s Stats) string {
	return fmt.Sprintf(`chop - token savings report

today: %d commands, %s tokens saved
total: %d commands, %s tokens saved (%.1f%% avg)

run 'chop gain --history' for command history`,
		s.TodayCommands, formatNum(s.TodaySavedTokens),
		s.TotalCommands, formatNum(s.TotalSavedTokens), s.OverallSavingsPct,
	)
}

// FormatHistory formats history records for display.
func FormatHistory(records []Record) string {
	if len(records) == 0 {
		return "no commands tracked yet"
	}
	var b strings.Builder
	b.WriteString("recent commands:\n")
	for _, r := range records {
		b.WriteString(fmt.Sprintf("  %s  %-20s %.1f%% saved (%d -> %d tokens)\n",
			r.Timestamp, r.Command, r.SavingsPct, r.RawTokens, r.FilteredTokens))
	}
	return b.String()
}

func formatNum(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("%d,%03d", n/1000, n%1000)
}
