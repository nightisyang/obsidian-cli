package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nightisyang/obsidian-cli/internal/app"
	"github.com/nightisyang/obsidian-cli/internal/errs"
	"github.com/nightisyang/obsidian-cli/internal/index"
	"github.com/nightisyang/obsidian-cli/internal/note"
)

func verifyHashPrecondition(rt *app.Runtime, path, expectedHash string) error {
	expected := strings.TrimSpace(expectedHash)
	if expected == "" {
		return nil
	}
	n, err := rt.Backend.GetNote(rt.Context, path)
	if err != nil {
		return err
	}
	actual := hashString(n.Raw)
	if actual != expected {
		return fmt.Errorf("hash precondition failed for %s: expected %s got %s", path, expected, actual)
	}
	return nil
}

func hashString(raw string) string {
	digest := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(digest[:])
}

func enforceWriteGuards(rt *app.Runtime, n note.Note) error {
	if err := enforceNoteSizeGuard(rt, n); err != nil {
		return err
	}
	if err := enforceBacklinkIntegrity(rt, n); err != nil {
		return err
	}
	if err := enforceNoOrphanNotes(rt, n); err != nil {
		return err
	}
	return nil
}

func enforceNoteSizeGuard(rt *app.Runtime, n note.Note) error {
	threshold := rootOpts.noteSizeMaxBytes
	if threshold <= 0 {
		return nil
	}
	sizeBytes := len([]byte(n.Raw))
	if sizeBytes <= threshold {
		return nil
	}
	splitPath := suggestedSplitPath(n.Path)
	splitTitle := suggestedSplitTitle(splitPath)
	baseRef := strings.TrimSuffix(path.Base(n.Path), ".md")
	splitRef := strings.TrimSuffix(path.Base(splitPath), ".md")
	message := fmt.Sprintf(
		"note %s is %d bytes, exceeding %d bytes. split content into %s and add reciprocal backlinks.\nexample:\n1) obsidian-cli --vault <vault> note create %q --content \"<moved section>\"\n2) obsidian-cli --vault <vault> note append %s \"Continued in [[%s]]\"\n3) obsidian-cli --vault <vault> note append %s \"Context from [[%s]]\"",
		n.Path,
		sizeBytes,
		threshold,
		splitPath,
		splitTitle,
		n.Path,
		splitRef,
		splitPath,
		baseRef,
	)
	return errs.NewDetailed(
		errs.ExitValidation,
		"note_size_limit_exceeded",
		"Split into a new note with reciprocal backlinks, or raise --note-size-max-bytes.",
		message,
	)
}

func suggestedSplitPath(currentPath string) string {
	clean := strings.TrimSpace(currentPath)
	if clean == "" {
		return "split-note-part-2.md"
	}
	if strings.HasSuffix(clean, ".md") {
		return strings.TrimSuffix(clean, ".md") + "-part-2.md"
	}
	return clean + "-part-2.md"
}

func suggestedSplitTitle(relPath string) string {
	base := path.Base(strings.TrimSpace(relPath))
	base = strings.TrimSuffix(base, ".md")
	base = strings.ReplaceAll(base, "-", " ")
	base = strings.TrimSpace(base)
	if base == "" {
		return "Split Note Part 2"
	}
	parts := strings.Fields(base)
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + strings.ToLower(parts[i][1:])
		}
	}
	return strings.Join(parts, " ")
}

func enforceBacklinkIntegrity(rt *app.Runtime, n note.Note) error {
	targets := index.ParseWikiLinks(n.Body)
	if len(targets) == 0 {
		return nil
	}
	files, err := index.ListMarkdownFiles(rt.VaultRoot)
	if err != nil {
		return err
	}
	paths := map[string]struct{}{}
	stems := map[string]struct{}{}
	for _, abs := range files {
		rel, relErr := filepath.Rel(rt.VaultRoot, abs)
		if relErr != nil {
			return relErr
		}
		rel = filepath.ToSlash(rel)
		pathKey := index.NormalizeLinkTarget(strings.TrimSuffix(rel, ".md"))
		stemKey := index.NormalizeLinkTarget(strings.TrimSuffix(filepath.Base(rel), ".md"))
		if pathKey != "" {
			paths[pathKey] = struct{}{}
		}
		if stemKey != "" {
			stems[stemKey] = struct{}{}
		}
	}

	unresolved := make([]string, 0)
	for _, target := range targets {
		if _, ok := paths[target]; ok {
			continue
		}
		if _, ok := stems[target]; ok {
			continue
		}
		unresolved = append(unresolved, target)
	}
	if len(unresolved) == 0 {
		return nil
	}
	sort.Strings(unresolved)
	first := unresolved[0]
	splitPath := first + ".md"
	currentStem := strings.TrimSuffix(path.Base(n.Path), ".md")
	message := fmt.Sprintf(
		"note %s has unresolved wikilinks: %s. ensure link targets exist so backlinks are valid.\nexample fix:\n1) obsidian-cli --vault <vault> note create %q\n2) obsidian-cli --vault <vault> note append %s \"See [[%s]]\"\n3) obsidian-cli --vault <vault> note append %s \"Referenced by [[%s]]\"",
		n.Path,
		strings.Join(unresolved, ", "),
		suggestedSplitTitle(splitPath),
		n.Path,
		first,
		splitPath,
		currentStem,
	)
	return errs.NewDetailed(
		errs.ExitValidation,
		"backlink_target_not_found",
		"Create missing target note(s) or correct wikilink path(s) before writing.",
		message,
	)
}

func enforceNoOrphanNotes(rt *app.Runtime, n note.Note) error {
	if !rootOpts.noOrphanNotes {
		return nil
	}

	idx, err := index.BuildIndex(rt.VaultRoot)
	if err != nil {
		return err
	}

	selfPath := index.NormalizeLinkTarget(strings.TrimSuffix(n.Path, ".md"))
	selfStem := index.NormalizeLinkTarget(strings.TrimSuffix(path.Base(n.Path), ".md"))

	existingPaths := map[string]struct{}{}
	existingStems := map[string]struct{}{}
	for rel := range idx.SourceToTarget {
		pathKey := index.NormalizeLinkTarget(strings.TrimSuffix(rel, ".md"))
		stemKey := index.NormalizeLinkTarget(strings.TrimSuffix(filepath.Base(rel), ".md"))
		if pathKey != "" {
			existingPaths[pathKey] = struct{}{}
		}
		if stemKey != "" {
			existingStems[stemKey] = struct{}{}
		}
	}

	hasOutgoing := false
	for _, target := range index.ParseWikiLinks(n.Body) {
		if target == selfPath || target == selfStem {
			continue
		}
		if _, ok := existingPaths[target]; ok {
			hasOutgoing = true
			break
		}
		if _, ok := existingStems[target]; ok {
			hasOutgoing = true
			break
		}
	}

	hasIncoming := false
	for _, source := range index.BacklinksForPath(rt.VaultRoot, idx, n.Path) {
		if source != n.Path {
			hasIncoming = true
			break
		}
	}

	if hasOutgoing || hasIncoming {
		return nil
	}

	stem := strings.TrimSuffix(path.Base(n.Path), ".md")
	message := fmt.Sprintf(
		"note %s has no graph connections to other notes (no inbound backlinks and no outgoing links). with --no-orphan-notes enabled this write is blocked.\nexample self-correction flow:\n1) obsidian-cli --vault <vault> note list --limit 1\n2) obsidian-cli --vault <vault> note append %s \"Related: [[existing-note]]\"\n3) obsidian-cli --vault <vault> note append existing-note.md \"Related: [[%s]]\"",
		n.Path,
		n.Path,
		stem,
	)
	return errs.NewDetailed(
		errs.ExitValidation,
		"orphan_note_not_allowed",
		"Link this note to an existing note and add a reciprocal backlink.",
		message,
	)
}
