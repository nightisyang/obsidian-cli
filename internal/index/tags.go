package index

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/nightisyang/obsidian-cli/internal/note"
	"github.com/nightisyang/obsidian-cli/internal/search"
)

type TagCount struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

type TagListOptions struct {
	Limit int
	Sort  string
}

var inlineTagPattern = regexp.MustCompile(`(?:^|\s)#([A-Za-z0-9_\-/]+)`)

func ExtractInlineTags(body string) []string {
	matches := inlineTagPattern.FindAllStringSubmatch(body, -1)
	seen := map[string]struct{}{}
	out := []string{}
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		tag := normalizeTag(m[1])
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	sort.Strings(out)
	return out
}

func AggregateTags(vaultRoot string) (map[string]int, error) {
	files, err := ListMarkdownFiles(vaultRoot)
	if err != nil {
		return nil, err
	}
	counts := map[string]int{}
	for _, abs := range files {
		rel, _ := filepath.Rel(vaultRoot, abs)
		rel = filepath.ToSlash(rel)
		n, readErr := note.Read(vaultRoot, rel)
		if readErr != nil {
			return nil, readErr
		}
		seen := map[string]struct{}{}
		for _, tag := range n.Frontmatter.Tags {
			norm := normalizeTag(tag)
			if norm == "" {
				continue
			}
			seen[norm] = struct{}{}
		}
		for _, tag := range ExtractInlineTags(n.Body) {
			seen[tag] = struct{}{}
		}
		for tag := range seen {
			counts[tag]++
		}
	}
	return counts, nil
}

func ListTags(vaultRoot string, opts TagListOptions) ([]TagCount, error) {
	counts, err := AggregateTags(vaultRoot)
	if err != nil {
		return nil, err
	}
	items := make([]TagCount, 0, len(counts))
	for tag, count := range counts {
		items = append(items, TagCount{Tag: tag, Count: count})
	}
	sortBy := opts.Sort
	if sortBy == "" {
		sortBy = "count"
	}
	sort.Slice(items, func(i, j int) bool {
		if sortBy == "name" {
			return items[i].Tag < items[j].Tag
		}
		if items[i].Count == items[j].Count {
			return items[i].Tag < items[j].Tag
		}
		return items[i].Count > items[j].Count
	})
	if opts.Limit > 0 && len(items) > opts.Limit {
		items = items[:opts.Limit]
	}
	return items, nil
}

func SearchTag(vaultRoot, tag string, limit int) ([]search.SearchResult, error) {
	norm := normalizeTag(tag)
	if norm == "" {
		return []search.SearchResult{}, nil
	}
	files, err := ListMarkdownFiles(vaultRoot)
	if err != nil {
		return nil, err
	}
	results := []search.SearchResult{}
	for _, abs := range files {
		rel, _ := filepath.Rel(vaultRoot, abs)
		rel = filepath.ToSlash(rel)
		n, readErr := note.Read(vaultRoot, rel)
		if readErr != nil {
			return nil, readErr
		}
		found := false
		for _, t := range n.Frontmatter.Tags {
			if normalizeTag(t) == norm {
				found = true
				break
			}
		}
		if !found {
			for _, t := range ExtractInlineTags(n.Body) {
				if t == norm {
					found = true
					break
				}
			}
		}
		if found {
			results = append(results, search.SearchResult{
				Path:      rel,
				Match:     "#" + norm,
				Snippet:   "tag match",
				MatchType: "tag",
			})
			if limit > 0 && len(results) >= limit {
				break
			}
		}
	}
	return results, nil
}

func ListMarkdownFiles(root string) ([]string, error) {
	files := []string{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") && path != root {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func normalizeTag(tag string) string {
	clean := strings.TrimSpace(strings.TrimPrefix(tag, "#"))
	return strings.ToLower(clean)
}
