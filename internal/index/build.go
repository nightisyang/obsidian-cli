package index

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/nightisyang/obsidian-cli/internal/note"
)

type BacklinkIndex struct {
	BuiltAt        time.Time           `json:"built_at"`
	SourceToTarget map[string][]string `json:"source_to_target"`
	TargetToSource map[string][]string `json:"target_to_source"`
	FileMTimeMax   time.Time           `json:"file_mtime_max"`
}

func BuildIndex(vaultRoot string) (BacklinkIndex, error) {
	files, err := ListMarkdownFiles(vaultRoot)
	if err != nil {
		return BacklinkIndex{}, err
	}
	idx := BacklinkIndex{
		BuiltAt:        time.Now().UTC(),
		SourceToTarget: map[string][]string{},
		TargetToSource: map[string][]string{},
	}
	maxMtime := time.Time{}

	for _, abs := range files {
		info, statErr := os.Stat(abs)
		if statErr == nil && info.ModTime().After(maxMtime) {
			maxMtime = info.ModTime()
		}
		rel, _ := filepath.Rel(vaultRoot, abs)
		rel = filepath.ToSlash(rel)
		n, readErr := note.Read(vaultRoot, rel)
		if readErr != nil {
			return BacklinkIndex{}, readErr
		}
		targets := ParseWikiLinks(n.Body)
		idx.SourceToTarget[rel] = targets
		for _, target := range targets {
			idx.TargetToSource[target] = append(idx.TargetToSource[target], rel)
		}
	}

	for key := range idx.TargetToSource {
		sort.Strings(idx.TargetToSource[key])
	}
	idx.FileMTimeMax = maxMtime
	return idx, nil
}

func BacklinksForPath(vaultRoot string, index BacklinkIndex, relPath string) []string {
	target := NormalizeLinkTarget(strings.TrimSuffix(relPath, ".md"))
	base := NormalizeLinkTarget(strings.TrimSuffix(filepath.Base(relPath), ".md"))
	combined := map[string]struct{}{}
	for _, source := range index.TargetToSource[target] {
		combined[source] = struct{}{}
	}
	for _, source := range index.TargetToSource[base] {
		combined[source] = struct{}{}
	}
	out := make([]string, 0, len(combined))
	for source := range combined {
		out = append(out, source)
	}
	sort.Strings(out)
	return out
}
