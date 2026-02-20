package note

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var wikilinkPattern = regexp.MustCompile(`\[\[([^\]]+)\]\]`)

func RewriteLinks(vaultRoot, oldRel, newRel string, dryRun bool) ([]string, error) {
	paths, err := listMarkdown(vaultRoot)
	if err != nil {
		return nil, err
	}

	oldKey := normalizeLinkKey(strings.TrimSuffix(filepath.ToSlash(oldRel), ".md"))
	oldStem := normalizeLinkKey(strings.TrimSuffix(filepath.Base(oldRel), ".md"))
	newKey := strings.TrimSuffix(filepath.ToSlash(newRel), ".md")
	changed := []string{}

	for _, abs := range paths {
		rel, _ := filepath.Rel(vaultRoot, abs)
		rel = filepath.ToSlash(rel)
		contentBytes, readErr := os.ReadFile(abs)
		if readErr != nil {
			return nil, readErr
		}
		content := string(contentBytes)
		updated := wikilinkPattern.ReplaceAllStringFunc(content, func(match string) string {
			inner := strings.TrimSuffix(strings.TrimPrefix(match, "[["), "]]")
			parts := strings.SplitN(inner, "|", 2)
			targetWithAnchor := strings.TrimSpace(parts[0])
			alias := ""
			if len(parts) == 2 {
				alias = parts[1]
			}

			target := targetWithAnchor
			anchor := ""
			if idx := strings.Index(targetWithAnchor, "#"); idx >= 0 {
				target = targetWithAnchor[:idx]
				anchor = targetWithAnchor[idx:]
			}

			norm := normalizeLinkKey(target)
			if norm != oldKey && norm != oldStem {
				return match
			}

			rewritten := newKey + anchor
			if alias != "" {
				return "[[" + rewritten + "|" + alias + "]]"
			}
			return "[[" + rewritten + "]]"
		})

		if updated != content {
			changed = append(changed, rel)
			if !dryRun {
				if writeErr := os.WriteFile(abs, []byte(updated), 0o644); writeErr != nil {
					return nil, writeErr
				}
			}
		}
	}
	return changed, nil
}

func normalizeLinkKey(target string) string {
	v := strings.TrimSpace(target)
	v = strings.TrimPrefix(v, "./")
	v = strings.TrimSuffix(v, ".md")
	v = filepath.ToSlash(v)
	return strings.ToLower(v)
}

func listMarkdown(root string) ([]string, error) {
	out := []string{}
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
			out = append(out, path)
		}
		return nil
	})
	return out, err
}
