//go:build windows

package tracking

import (
	"io"
	"os"
	"path/filepath"
)

// MigrateWindowsDataDir migrates tracking data from the legacy path
// (~/.local/share/chop) to the current Windows data dir (%LocalAppData%\chop).
//
// It runs only when:
//   - the legacy DB exists
//   - the new DB is absent or contains only schema (size <= 8192 bytes)
//   - the sentinel file ~/.local/share/chop/.migrated does not exist
//
// Silent on all errors — a failed migration is not fatal.
func MigrateWindowsDataDir() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	legacyDir := filepath.Join(home, ".local", "share", "chop")
	sentinelPath := filepath.Join(legacyDir, ".migrated")

	// Already migrated.
	if _, err := os.Stat(sentinelPath); err == nil {
		return
	}

	legacyDB := filepath.Join(legacyDir, "tracking.db")
	if _, err := os.Stat(legacyDB); os.IsNotExist(err) {
		// Nothing to migrate.
		return
	}

	// Determine new data dir from dbPath() so we respect CHOP_DB_PATH in tests.
	newDB := dbPath()
	newDir := filepath.Dir(newDB)

	// If the new DB already has real data, don't overwrite it.
	if fi, err := os.Stat(newDB); err == nil && fi.Size() > 8192 {
		// Write sentinel so we don't check again.
		_ = writeSentinel(sentinelPath)
		return
	}

	if err := os.MkdirAll(newDir, 0o700); err != nil {
		return
	}

	// Files to migrate: DB plus optional WAL/SHM and hook audit log.
	candidates := []string{
		"tracking.db",
		"tracking.db-wal",
		"tracking.db-shm",
		"hook-audit.log",
	}

	for _, name := range candidates {
		src := filepath.Join(legacyDir, name)
		dst := filepath.Join(newDir, name)
		_ = copyFile(src, dst)
	}

	_ = writeSentinel(sentinelPath)
}

func copyFile(src, dst string) error {
	sf, err := os.Open(src)
	if err != nil {
		return err // source absent — fine
	}
	defer sf.Close()

	df, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer df.Close()

	_, err = io.Copy(df, sf)
	return err
}

func writeSentinel(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	return f.Close()
}
