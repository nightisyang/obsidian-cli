package backend

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/nightisyang/obsidian-cli/internal/errs"
	"github.com/nightisyang/obsidian-cli/internal/frontmatter"
	"github.com/nightisyang/obsidian-cli/internal/index"
	"github.com/nightisyang/obsidian-cli/internal/note"
	"github.com/nightisyang/obsidian-cli/internal/search"
	"github.com/nightisyang/obsidian-cli/internal/vault"
)

type NativeBackend struct {
	vaultRoot string
	cfg       vault.Config
	mode      string
	engine    search.Engine
}

func NewNativeBackend(vaultRoot string, cfg vault.Config, mode string) *NativeBackend {
	return &NativeBackend{
		vaultRoot: vaultRoot,
		cfg:       cfg,
		mode:      mode,
		engine:    &search.RGEngine{VaultRoot: vaultRoot},
	}
}

func (b *NativeBackend) VaultStatus(_ context.Context) (vault.Status, error) {
	return vault.ComputeStatus(b.vaultRoot, "", "", b.mode)
}

func (b *NativeBackend) CreateNote(_ context.Context, in note.CreateInput) (note.Note, error) {
	return note.Create(b.vaultRoot, in)
}

func (b *NativeBackend) GetNote(_ context.Context, path string) (note.Note, error) {
	return note.Get(b.vaultRoot, path)
}

func (b *NativeBackend) AppendNote(_ context.Context, path, content string) (note.Note, error) {
	return note.Append(b.vaultRoot, path, content)
}

func (b *NativeBackend) PrependNote(_ context.Context, path, content string) (note.Note, error) {
	return note.Prepend(b.vaultRoot, path, content)
}

func (b *NativeBackend) DeleteNote(_ context.Context, path string) error {
	return note.Delete(b.vaultRoot, path)
}

func (b *NativeBackend) ListNotes(_ context.Context, dir string, opts note.ListOptions) ([]note.Note, error) {
	return note.List(b.vaultRoot, dir, opts)
}

func (b *NativeBackend) MoveNote(_ context.Context, src, dst string, opts note.MoveOptions) (note.Note, error) {
	n, _, err := note.Move(b.vaultRoot, src, dst, opts)
	return n, err
}

func (b *NativeBackend) Search(ctx context.Context, q search.Query) ([]search.SearchResult, error) {
	switch q.Type {
	case search.QueryText:
		return b.engine.Search(ctx, q)
	case search.QueryTag:
		results, err := index.SearchTag(b.vaultRoot, q.Tag, q.Limit)
		if err != nil {
			return nil, err
		}
		return filterByPath(results, q.Path), nil
	case search.QueryProp:
		return b.searchByProp(q)
	default:
		return nil, errs.New(errs.ExitValidation, "unknown search query type")
	}
}

func (b *NativeBackend) ListTags(_ context.Context, opts index.TagListOptions) ([]index.TagCount, error) {
	return index.ListTags(b.vaultRoot, opts)
}

func (b *NativeBackend) SearchTag(_ context.Context, tag string, limit int) ([]search.SearchResult, error) {
	return index.SearchTag(b.vaultRoot, tag, limit)
}

func (b *NativeBackend) OutgoingLinks(_ context.Context, path string) ([]string, error) {
	n, err := note.Get(b.vaultRoot, path)
	if err != nil {
		return nil, err
	}
	return index.ParseWikiLinks(n.Body), nil
}

func (b *NativeBackend) Backlinks(_ context.Context, path string, rebuild bool) ([]string, error) {
	var idx index.BacklinkIndex
	var ok bool
	if !rebuild {
		idx, ok = index.GetCached(b.vaultRoot)
	}
	if !ok || rebuild {
		built, err := index.BuildIndex(b.vaultRoot)
		if err != nil {
			return nil, err
		}
		idx = built
		index.SetCached(b.vaultRoot, idx)
	}
	_, rel, err := vault.ResolveNoteAbs(b.vaultRoot, path)
	if err != nil {
		return nil, err
	}
	return index.BacklinksForPath(b.vaultRoot, idx, rel), nil
}

func (b *NativeBackend) PropGet(_ context.Context, path, key string) (any, error) {
	n, err := note.Get(b.vaultRoot, path)
	if err != nil {
		return nil, err
	}
	values := frontmatter.FrontmatterToMap(n.Frontmatter)
	value, ok := values[key]
	if !ok {
		return nil, errs.New(errs.ExitNotFound, "property not found")
	}
	return value, nil
}

func (b *NativeBackend) PropSet(_ context.Context, path, key string, value any) (note.Note, error) {
	n, err := note.Get(b.vaultRoot, path)
	if err != nil {
		return note.Note{}, err
	}
	values := frontmatter.FrontmatterToMap(n.Frontmatter)
	values[key] = value
	n.Frontmatter = frontmatter.MapToFrontmatter(values)
	return note.Write(b.vaultRoot, n.Path, n, false, now())
}

func (b *NativeBackend) PropList(_ context.Context, path string) (map[string]any, error) {
	n, err := note.Get(b.vaultRoot, path)
	if err != nil {
		return nil, err
	}
	return frontmatter.FrontmatterToMap(n.Frontmatter), nil
}

func (b *NativeBackend) searchByProp(q search.Query) ([]search.SearchResult, error) {
	notes, err := note.List(b.vaultRoot, q.Path, note.ListOptions{Recursive: true})
	if err != nil {
		return nil, err
	}
	matches := []search.SearchResult{}
	for _, n := range notes {
		props := frontmatter.FrontmatterToMap(n.Frontmatter)
		value, ok := props[q.PropKey]
		if !ok {
			continue
		}
		if strings.EqualFold(fmt.Sprint(value), q.PropValue) {
			matches = append(matches, search.SearchResult{
				Path:      n.Path,
				Match:     fmt.Sprintf("%s=%v", q.PropKey, value),
				Snippet:   "frontmatter property match",
				MatchType: "prop",
			})
			if q.Limit > 0 && len(matches) >= q.Limit {
				break
			}
		}
	}
	return matches, nil
}

func filterByPath(results []search.SearchResult, prefix string) []search.SearchResult {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return results
	}
	clean := filepath.ToSlash(strings.TrimPrefix(prefix, "/"))
	out := []search.SearchResult{}
	for _, r := range results {
		if strings.HasPrefix(r.Path, clean) {
			out = append(out, r)
		}
	}
	return out
}
