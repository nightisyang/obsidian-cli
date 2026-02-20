package cmd

import (
	"strings"
	"time"

	"github.com/nightisyang/obsidian-cli/internal/errs"
)

func parseDateFlag(raw string) (time.Time, error) {
	clean := strings.TrimSpace(raw)
	if clean == "" {
		return time.Now(), nil
	}
	t, err := time.Parse("2006-01-02", clean)
	if err != nil {
		return time.Time{}, errs.New(errs.ExitValidation, "date must use YYYY-MM-DD")
	}
	return t, nil
}
