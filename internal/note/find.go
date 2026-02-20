package note

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/nightisyang/obsidian-cli/internal/frontmatter"
)

type FindFilters struct {
	Kind   string
	Tags   []string
	Status string
	Topic  string
	Since  *time.Time
}

type FindResult struct {
	Path      string   `json:"path"`
	Kind      string   `json:"kind"`
	Tags      []string `json:"tags"`
	Status    string   `json:"status"`
	Topic     string   `json:"topic"`
	CreatedAt string   `json:"created_at"`
}

func FindByMetadata(vaultRoot string, filters FindFilters) ([]FindResult, error) {
	filters = normalizeFindFilters(filters)
	results := []FindResult{}

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

		values, _, _, err := frontmatter.ParseDocument(string(payload))
		if err != nil {
			return err
		}
		fm := frontmatter.MapToFrontmatter(values)
		createdAt := extractCreatedAt(values, fm)
		if !matchesFindFilters(fm, createdAt, filters) {
			return nil
		}

		rel, err := filepath.Rel(vaultRoot, path)
		if err != nil {
			return err
		}

		item := FindResult{
			Path:   filepath.ToSlash(rel),
			Kind:   fm.Kind,
			Tags:   append([]string(nil), fm.Tags...),
			Status: fm.Status,
			Topic:  fm.Topic,
		}
		if createdAt != nil {
			item.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		}
		results = append(results, item)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Path < results[j].Path
	})
	return results, nil
}

func normalizeFindFilters(filters FindFilters) FindFilters {
	filters.Kind = strings.TrimSpace(filters.Kind)
	filters.Status = strings.TrimSpace(filters.Status)
	filters.Topic = strings.TrimSpace(filters.Topic)
	filters.Tags = normalizeTags(filters.Tags)
	if filters.Since != nil {
		t := filters.Since.UTC()
		filters.Since = &t
	}
	return filters
}

func normalizeTags(raw []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(raw))
	for _, value := range raw {
		tag := strings.TrimSpace(strings.TrimPrefix(value, "#"))
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	return out
}

func extractCreatedAt(values map[string]any, fm frontmatter.Data) *time.Time {
	if fm.CreatedAt != nil {
		t := fm.CreatedAt.UTC()
		return &t
	}

	value, ok := values["created_at"]
	if !ok {
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		t := v.UTC()
		return &t
	case string:
		return parseFlexibleTime(v)
	default:
		return nil
	}
}

func matchesFindFilters(fm frontmatter.Data, createdAt *time.Time, filters FindFilters) bool {
	if filters.Kind != "" && fm.Kind != filters.Kind {
		return false
	}
	if filters.Status != "" && fm.Status != filters.Status {
		return false
	}
	if filters.Topic != "" && fm.Topic != filters.Topic {
		return false
	}
	if len(filters.Tags) > 0 {
		noteTags := map[string]struct{}{}
		for _, tag := range normalizeTags(fm.Tags) {
			noteTags[tag] = struct{}{}
		}
		for _, tag := range filters.Tags {
			if _, ok := noteTags[tag]; !ok {
				return false
			}
		}
	}
	if filters.Since != nil {
		if createdAt == nil {
			return false
		}
		if createdAt.Before(*filters.Since) {
			return false
		}
	}
	return true
}

func parseFlexibleTime(raw string) *time.Time {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		u := t.UTC()
		return &u
	}
	if t, err := time.Parse("2006-01-02", value); err == nil {
		u := t.UTC()
		return &u
	}
	return nil
}
