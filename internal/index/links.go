package index

import (
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var wikilinkRe = regexp.MustCompile(`\[\[([^\]]+)\]\]`)

func ParseWikiLinks(body string) []string {
	matches := wikilinkRe.FindAllStringSubmatch(body, -1)
	seen := map[string]struct{}{}
	out := []string{}
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		inner := strings.TrimSpace(m[1])
		parts := strings.SplitN(inner, "|", 2)
		target := strings.TrimSpace(parts[0])
		if idx := strings.Index(target, "#"); idx >= 0 {
			target = target[:idx]
		}
		norm := NormalizeLinkTarget(target)
		if norm == "" {
			continue
		}
		if _, ok := seen[norm]; ok {
			continue
		}
		seen[norm] = struct{}{}
		out = append(out, norm)
	}
	sort.Strings(out)
	return out
}

func NormalizeLinkTarget(target string) string {
	v := strings.TrimSpace(target)
	if v == "" {
		return ""
	}
	v = strings.TrimPrefix(v, "./")
	v = strings.TrimSuffix(v, ".md")
	v = filepath.ToSlash(filepath.Clean(v))
	v = strings.TrimPrefix(v, "../")
	v = strings.TrimPrefix(v, "/")
	return strings.ToLower(v)
}
