package note

import (
	"os"
	"path/filepath"
	"time"

	"github.com/nightisyang/obsidian-cli/internal/frontmatter"
)

func Write(vaultRoot, relPath string, n Note, creating bool, now time.Time) (Note, error) {
	abs, normalized, err := resolveNoteAbs(vaultRoot, relPath)
	if err != nil {
		return Note{}, err
	}

	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return Note{}, err
	}

	applyTimestamps(&n.Frontmatter, creating, now)
	rendered, err := frontmatter.RenderMarkdown(n.Frontmatter, n.Body)
	if err != nil {
		return Note{}, err
	}

	tmp := abs + ".tmp"
	if err := os.WriteFile(tmp, []byte(rendered), 0o644); err != nil {
		return Note{}, err
	}
	if err := os.Rename(tmp, abs); err != nil {
		return Note{}, err
	}

	n.Path = normalized
	n.Title = titleFromPath(normalized)
	n.Raw = rendered
	return n, nil
}

func applyTimestamps(fm *Frontmatter, creating bool, now time.Time) {
	if fm == nil {
		return
	}
	t := now.UTC()
	if creating && fm.CreatedAt == nil {
		fm.CreatedAt = &t
	}
	fm.UpdatedAt = &t
}
