package main

import (
	"fmt"
	"strings"
	"path/filepath"
)

func sanitizeFilename(s string) string {
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.ReplaceAll(s, "..", "__")
	return s
}

func main() {
    baseName := sanitizeFilename(".../abc")
    fmt.Println(baseName)

    // what about ".a." -> ".a."
    baseName2 := sanitizeFilename(".../../abc")
    fmt.Println(baseName2)

    // Try bypassing by putting separator inside ... ?
    // Windows supports / and \ and \x5c \x2f
    baseName3 := sanitizeFilename(".\\./")
    fmt.Println(baseName3)

    // Using unicode homoglyphs? Not supported by Windows filesystem driver directly to escape.

    fmt.Println(filepath.Join("tests", "fixtures", sanitizeFilename("..") + ".txt"))
    fmt.Println(filepath.Join("tests", "fixtures", sanitizeFilename("../..") + ".txt"))
}
