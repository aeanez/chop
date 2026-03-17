//go:build !windows

package tracking

// MigrateWindowsDataDir is a no-op on non-Windows platforms.
func MigrateWindowsDataDir() {}
