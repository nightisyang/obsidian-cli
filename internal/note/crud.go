package note

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/nightisyang/obsidian-cli/internal/errs"
)

func Create(vaultRoot string, in CreateInput) (Note, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return Note{}, errs.New(errs.ExitValidation, "title is required")
	}
	dir := strings.Trim(strings.TrimSpace(in.Dir), "/")
	slug := slugify(title)
	rel := slug + ".md"
	if dir != "" {
		rel = filepath.ToSlash(filepath.Join(dir, rel))
	}

	abs, normalized, err := resolveNoteAbs(vaultRoot, rel)
	if err != nil {
		return Note{}, err
	}
	if _, err := os.Stat(abs); err == nil {
		return Note{}, errs.New(errs.ExitValidation, "note already exists")
	}

	n := Note{
		Path:  normalized,
		Title: title,
		Frontmatter: Frontmatter{
			Tags:   dedupeStrings(in.Tags),
			Kind:   in.Kind,
			Topic:  in.Topic,
			Status: in.Status,
			Extra:  map[string]any{},
		},
		Body: in.Content,
	}
	return Write(vaultRoot, normalized, n, true, time.Now())
}

func Get(vaultRoot, path string) (Note, error) {
	return Read(vaultRoot, path)
}

func List(vaultRoot, dir string, opts ListOptions) ([]Note, error) {
	root := vaultRoot
	if strings.TrimSpace(dir) != "" {
		abs, _, err := resolveNoteAbs(vaultRoot, dir+"/placeholder.md")
		if err != nil {
			return nil, err
		}
		root = filepath.Dir(abs)
	}
	entries := []Note{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if !opts.Recursive && path != root {
				return filepath.SkipDir
			}
			if strings.HasPrefix(d.Name(), ".") && path != root {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}
		rel, _ := filepath.Rel(vaultRoot, path)
		rel = filepath.ToSlash(rel)
		n, readErr := Read(vaultRoot, rel)
		if readErr != nil {
			return readErr
		}
		entries = append(entries, n)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sortBy := opts.Sort
	if sortBy == "" {
		sortBy = "name"
	}
	sort.Slice(entries, func(i, j int) bool {
		switch sortBy {
		case "created":
			ai := time.Time{}
			aj := time.Time{}
			if entries[i].Frontmatter.CreatedAt != nil {
				ai = *entries[i].Frontmatter.CreatedAt
			}
			if entries[j].Frontmatter.CreatedAt != nil {
				aj = *entries[j].Frontmatter.CreatedAt
			}
			return ai.After(aj)
		case "updated":
			ai := time.Time{}
			aj := time.Time{}
			if entries[i].Frontmatter.UpdatedAt != nil {
				ai = *entries[i].Frontmatter.UpdatedAt
			}
			if entries[j].Frontmatter.UpdatedAt != nil {
				aj = *entries[j].Frontmatter.UpdatedAt
			}
			return ai.After(aj)
		default:
			return entries[i].Path < entries[j].Path
		}
	})

	if opts.Limit > 0 && len(entries) > opts.Limit {
		entries = entries[:opts.Limit]
	}
	return entries, nil
}

func Delete(vaultRoot, path string) error {
	abs, _, err := resolveNoteAbs(vaultRoot, path)
	if err != nil {
		return err
	}
	if err := os.Remove(abs); err != nil {
		if os.IsNotExist(err) {
			return errs.New(errs.ExitNotFound, "note not found")
		}
		return err
	}
	return nil
}

func Append(vaultRoot, path, content string) (Note, error) {
	return AppendWithOptions(vaultRoot, path, content, false)
}

func AppendWithOptions(vaultRoot, path, content string, inline bool) (Note, error) {
	n, err := Read(vaultRoot, path)
	if err != nil {
		return Note{}, err
	}
	text := content
	if !inline && n.Body != "" && !strings.HasSuffix(n.Body, "\n") {
		n.Body += "\n"
	}
	n.Body += text
	return Write(vaultRoot, n.Path, n, false, time.Now())
}

func Prepend(vaultRoot, path, content string) (Note, error) {
	return PrependWithOptions(vaultRoot, path, content, false)
}

func PrependWithOptions(vaultRoot, path, content string, inline bool) (Note, error) {
	n, err := Read(vaultRoot, path)
	if err != nil {
		return Note{}, err
	}
	text := content
	if !inline && text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	n.Body = text + n.Body
	return Write(vaultRoot, n.Path, n, false, time.Now())
}

func Move(vaultRoot, src, dst string, opts MoveOptions) (Note, []string, error) {
	srcAbs, srcNorm, err := resolveNoteAbs(vaultRoot, src)
	if err != nil {
		return Note{}, nil, err
	}
	dstAbs, dstNorm, err := resolveNoteAbs(vaultRoot, dst)
	if err != nil {
		return Note{}, nil, err
	}
	if _, statErr := os.Stat(srcAbs); statErr != nil {
		if os.IsNotExist(statErr) {
			return Note{}, nil, errs.New(errs.ExitNotFound, "source note not found")
		}
		return Note{}, nil, statErr
	}
	if _, statErr := os.Stat(dstAbs); statErr == nil {
		return Note{}, nil, errs.New(errs.ExitValidation, "destination already exists")
	}

	rewritten := []string{}
	if opts.DryRun {
		n, readErr := Read(vaultRoot, srcNorm)
		if readErr != nil {
			return Note{}, nil, readErr
		}
		n.Path = dstNorm
		n.Title = titleFromPath(dstNorm)
		if opts.UpdateLinks {
			rewritten, _ = RewriteLinks(vaultRoot, srcNorm, dstNorm, true)
		}
		return n, rewritten, nil
	}

	if err := os.MkdirAll(filepath.Dir(dstAbs), 0o755); err != nil {
		return Note{}, nil, err
	}
	if err := os.Rename(srcAbs, dstAbs); err != nil {
		return Note{}, nil, err
	}

	n, err := Read(vaultRoot, dstNorm)
	if err != nil {
		return Note{}, nil, err
	}
	n, err = Write(vaultRoot, dstNorm, n, false, time.Now())
	if err != nil {
		return Note{}, nil, err
	}

	if opts.UpdateLinks {
		rewritten, err = RewriteLinks(vaultRoot, srcNorm, dstNorm, false)
		if err != nil {
			return Note{}, nil, err
		}
	}
	return n, rewritten, nil
}

func slugify(input string) string {
	lower := strings.ToLower(strings.TrimSpace(input))
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug := re.ReplaceAllString(lower, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		return "untitled"
	}
	return slug
}

func dedupeStrings(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, value := range in {
		v := strings.TrimSpace(strings.TrimPrefix(value, "#"))
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
