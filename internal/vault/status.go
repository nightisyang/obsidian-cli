package vault

import (
	"path/filepath"

	"github.com/nightisyang/obsidian-cli/internal/index"
)

type Status struct {
	Root          string `json:"root"`
	ConfigPath    string `json:"config_path,omitempty"`
	ConfigSource  string `json:"config_source"`
	EffectiveMode string `json:"effective_mode"`
	NoteCount     int    `json:"note_count"`
	TagCount      int    `json:"tag_count"`
}

func ComputeStatus(vaultRoot, configPath, source, effectiveMode string) (Status, error) {
	notes, err := filepath.Glob(filepath.Join(vaultRoot, "**", "*.md"))
	if err != nil || len(notes) == 0 {
		// filepath.Glob does not support **; fallback to walker in index package
		notes, err = index.ListMarkdownFiles(vaultRoot)
		if err != nil {
			return Status{}, err
		}
	}

	tags, err := index.AggregateTags(vaultRoot)
	if err != nil {
		return Status{}, err
	}

	return Status{
		Root:          vaultRoot,
		ConfigPath:    configPath,
		ConfigSource:  source,
		EffectiveMode: effectiveMode,
		NoteCount:     len(notes),
		TagCount:      len(tags),
	}, nil
}
