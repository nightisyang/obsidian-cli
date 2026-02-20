package note

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/nightisyang/obsidian-cli/internal/errs"
	"github.com/nightisyang/obsidian-cli/internal/frontmatter"
)

func Read(vaultRoot, relPath string) (Note, error) {
	abs, normalized, err := resolveNoteAbs(vaultRoot, relPath)
	if err != nil {
		return Note{}, err
	}
	content, err := os.ReadFile(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return Note{}, errs.New(errs.ExitNotFound, "note not found")
		}
		return Note{}, err
	}

	values, body, _, err := frontmatter.ParseDocument(string(content))
	if err != nil {
		return Note{}, errs.Wrap(errs.ExitValidation, "failed to parse frontmatter", err)
	}
	fm := frontmatter.MapToFrontmatter(values)
	return Note{
		Path:        normalized,
		Title:       titleFromPath(normalized),
		Frontmatter: fm,
		Body:        body,
		Raw:         string(content),
	}, nil
}

func titleFromPath(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}
