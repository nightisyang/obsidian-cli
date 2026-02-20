package note

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nightisyang/obsidian-cli/internal/errs"
)

type dailyNotesConfig struct {
	Folder string `json:"folder"`
	Format string `json:"format"`
}

func ResolveDailyPath(vaultRoot string, at time.Time) (string, error) {
	cfg := dailyNotesConfig{Format: "YYYY-MM-DD"}
	path := filepath.Join(vaultRoot, ".obsidian", "daily-notes.json")
	payload, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(payload, &cfg)
	}

	if strings.TrimSpace(cfg.Format) == "" {
		cfg.Format = "YYYY-MM-DD"
	}
	layout := momentToGoLayout(cfg.Format)
	name := at.Format(layout) + ".md"

	folder := strings.Trim(strings.ReplaceAll(cfg.Folder, "\\", "/"), "/")
	if folder == "" {
		return filepath.ToSlash(name), nil
	}
	return filepath.ToSlash(filepath.Join(folder, name)), nil
}

func EnsureExists(vaultRoot, path string) (Note, error) {
	abs, normalized, err := resolveNoteAbs(vaultRoot, path)
	if err != nil {
		return Note{}, err
	}
	if _, statErr := os.Stat(abs); statErr == nil {
		return Read(vaultRoot, normalized)
	}

	if mkErr := os.MkdirAll(filepath.Dir(abs), 0o755); mkErr != nil {
		return Note{}, mkErr
	}
	if writeErr := os.WriteFile(abs, nil, 0o644); writeErr != nil {
		return Note{}, writeErr
	}
	return Read(vaultRoot, normalized)
}

func DailyRead(vaultRoot string, at time.Time, create bool) (Note, string, error) {
	path, err := ResolveDailyPath(vaultRoot, at)
	if err != nil {
		return Note{}, "", err
	}

	if create {
		n, createErr := EnsureExists(vaultRoot, path)
		return n, path, createErr
	}

	n, readErr := Read(vaultRoot, path)
	if readErr != nil {
		return Note{}, "", readErr
	}
	return n, path, nil
}

func DailyAppend(vaultRoot string, at time.Time, content string, inline bool) (Note, string, error) {
	path, err := ResolveDailyPath(vaultRoot, at)
	if err != nil {
		return Note{}, "", err
	}
	if _, err := EnsureExists(vaultRoot, path); err != nil {
		return Note{}, "", err
	}
	n, err := AppendWithOptions(vaultRoot, path, content, inline)
	return n, path, err
}

func DailyPrepend(vaultRoot string, at time.Time, content string, inline bool) (Note, string, error) {
	path, err := ResolveDailyPath(vaultRoot, at)
	if err != nil {
		return Note{}, "", err
	}
	if _, err := EnsureExists(vaultRoot, path); err != nil {
		return Note{}, "", err
	}
	n, err := PrependWithOptions(vaultRoot, path, content, inline)
	return n, path, err
}

func momentToGoLayout(format string) string {
	replacer := strings.NewReplacer(
		"YYYY", "2006",
		"YY", "06",
		"MM", "01",
		"DD", "02",
		"HH", "15",
		"hh", "03",
		"mm", "04",
		"ss", "05",
	)
	layout := replacer.Replace(format)
	if strings.TrimSpace(layout) == "" {
		return "2006-01-02"
	}
	// If no supported token was present, use a stable fallback.
	if layout == format {
		return "2006-01-02"
	}
	return layout
}

func MustDailyPath(vaultRoot string, at time.Time) string {
	path, err := ResolveDailyPath(vaultRoot, at)
	if err != nil {
		return at.Format("2006-01-02") + ".md"
	}
	return path
}

func ValidateNoMultilineContent(content string) error {
	if strings.Contains(content, "\n") {
		return errs.New(errs.ExitValidation, "content must be a single line")
	}
	return nil
}
