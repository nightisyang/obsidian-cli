package search

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/nightisyang/obsidian-cli/internal/errs"
)

// StdlibEngine is a fallback text search engine using standard library only.
// It is used when ripgrep is not available on the system.
type StdlibEngine struct {
	VaultRoot string
}

func (e *StdlibEngine) Search(ctx context.Context, q Query) ([]SearchResult, error) {
	if q.Type != QueryText {
		return nil, errs.New(errs.ExitValidation, "stdlib engine supports text queries only")
	}

	searchRoot := filepath.Clean(e.VaultRoot)
	if strings.TrimSpace(q.Path) != "" {
		candidate := filepath.Clean(filepath.Join(searchRoot, filepath.FromSlash(q.Path)))
		if candidate != searchRoot && !strings.HasPrefix(candidate, searchRoot+string(filepath.Separator)) {
			return nil, errs.New(errs.ExitValidation, "search path escapes vault root")
		}
		searchRoot = candidate
	}

	needle := q.Text
	if !q.CaseSensitive {
		needle = strings.ToLower(needle)
	}

	var results []SearchResult
	err := filepath.WalkDir(searchRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil // skip unreadable dirs
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") && path != searchRoot {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		content := string(data)
		compare := content
		if !q.CaseSensitive {
			compare = strings.ToLower(content)
		}

		if !strings.Contains(compare, needle) {
			return nil
		}

		rel, _ := filepath.Rel(e.VaultRoot, path)
		rel = filepath.ToSlash(rel)

		// Find matching lines for snippet.
		lines := strings.Split(content, "\n")
		compareLines := strings.Split(compare, "\n")
		for lineNum, line := range compareLines {
			if !strings.Contains(line, needle) {
				continue
			}
			snippet := strings.TrimSpace(lines[lineNum])
			if len(snippet) > 120 {
				snippet = snippet[:120] + "…"
			}
			results = append(results, SearchResult{
				Path:    rel,
				Match:   snippet,
				Snippet: snippet,
				Line:    lineNum + 1,
			})
			if q.Limit > 0 && len(results) >= q.Limit {
				return filepath.SkipAll
			}
		}
		return nil
	})
	if err != nil {
		return nil, errs.Wrap(errs.ExitGeneric, "search failed", err)
	}
	return results, nil
}
