package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/nightisyang/obsidian-cli/internal/frontmatter"
)

type MigrateOptions struct {
	DryRun bool
	Kind   string
}

type MigrationFile struct {
	Path   string `json:"path"`
	Action string `json:"action"`
}

type MigrationResult struct {
	Migrated int             `json:"migrated"`
	Skipped  int             `json:"skipped"`
	Files    []MigrationFile `json:"files"`
}

func MigrateFrontmatter(vaultRoot string, opts MigrateOptions) (MigrationResult, error) {
	kind := strings.TrimSpace(opts.Kind)
	if kind == "" {
		kind = "note"
	}

	result := MigrationResult{Files: []MigrationFile{}}
	err := filepath.WalkDir(vaultRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") && path != vaultRoot {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}

		payload, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		_, _, hasFrontmatter, err := frontmatter.ParseDocument(string(payload))
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(vaultRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		if hasFrontmatter {
			result.Skipped++
			result.Files = append(result.Files, MigrationFile{
				Path:   rel,
				Action: "skipped",
			})
			return nil
		}

		if opts.DryRun {
			result.Files = append(result.Files, MigrationFile{
				Path:   rel,
				Action: "would_migrate",
			})
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		createdAt := info.ModTime().UTC().Format(time.RFC3339)
		updated := migrationHeader(kind, createdAt) + string(payload)
		if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
			return err
		}

		result.Migrated++
		result.Files = append(result.Files, MigrationFile{
			Path:   rel,
			Action: "migrated",
		})
		return nil
	})
	if err != nil {
		return MigrationResult{}, err
	}

	sort.Slice(result.Files, func(i, j int) bool {
		return result.Files[i].Path < result.Files[j].Path
	})
	return result, nil
}

func migrationHeader(kind, createdAt string) string {
	return fmt.Sprintf("---\nkind: %s\ntags: []\ncreated_at: %s\n---\n\n", kind, createdAt)
}
