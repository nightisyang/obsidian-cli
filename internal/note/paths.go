package note

import (
	"path/filepath"
	"strings"

	"github.com/nightisyang/obsidian-cli/internal/errs"
)

func normalizeNotePath(path string) string {
	trimmed := strings.TrimSpace(path)
	trimmed = strings.ReplaceAll(trimmed, "\\", "/")
	trimmed = strings.TrimPrefix(trimmed, "/")
	if !strings.HasSuffix(strings.ToLower(trimmed), ".md") {
		trimmed += ".md"
	}
	return filepath.ToSlash(filepath.Clean(trimmed))
}

func resolveNoteAbs(vaultRoot, relPath string) (string, string, error) {
	normalized := normalizeNotePath(relPath)
	abs := filepath.Join(vaultRoot, normalized)
	cleanAbs := filepath.Clean(abs)
	cleanRoot := filepath.Clean(vaultRoot)
	if cleanAbs != cleanRoot && !strings.HasPrefix(cleanAbs, cleanRoot+string(filepath.Separator)) {
		return "", "", errs.New(errs.ExitValidation, "path escapes vault root")
	}
	return cleanAbs, normalized, nil
}
