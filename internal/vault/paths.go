package vault

import (
	"path/filepath"
	"strings"

	"github.com/nightisyang/obsidian-cli/internal/errs"
)

func NormalizeNotePath(path string) string {
	trimmed := strings.TrimSpace(path)
	trimmed = strings.ReplaceAll(trimmed, "\\", "/")
	trimmed = strings.TrimPrefix(trimmed, "/")
	if !strings.HasSuffix(strings.ToLower(trimmed), ".md") {
		trimmed += ".md"
	}
	return filepath.ToSlash(filepath.Clean(trimmed))
}

func ResolveNoteAbs(vaultRoot, relPath string) (string, string, error) {
	normalized := NormalizeNotePath(relPath)
	abs := filepath.Join(vaultRoot, normalized)
	cleanAbs := filepath.Clean(abs)
	cleanRoot := filepath.Clean(vaultRoot)
	if cleanAbs != cleanRoot && !strings.HasPrefix(cleanAbs, cleanRoot+string(filepath.Separator)) {
		return "", "", errs.New(errs.ExitValidation, "path escapes vault root")
	}
	return cleanAbs, normalized, nil
}
