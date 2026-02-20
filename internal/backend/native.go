package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/nightisyang/obsidian-cli/internal/errs"
	"github.com/nightisyang/obsidian-cli/internal/frontmatter"
	"github.com/nightisyang/obsidian-cli/internal/index"
	"github.com/nightisyang/obsidian-cli/internal/note"
	"github.com/nightisyang/obsidian-cli/internal/search"
	"github.com/nightisyang/obsidian-cli/internal/tasks"
	"github.com/nightisyang/obsidian-cli/internal/templates"
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
	if strings.TrimSpace(in.Template) != "" {
		tpl, err := templates.Read(b.vaultRoot, b.cfg, in.Template, in.Title, true, now())
		if err != nil {
			return note.Note{}, err
		}
		in.Content = tpl.Content
	}
	return note.Create(b.vaultRoot, in)
}

func (b *NativeBackend) GetNote(_ context.Context, path string) (note.Note, error) {
	return note.Get(b.vaultRoot, path)
}

func (b *NativeBackend) GetHeading(_ context.Context, path, heading string) (note.HeadingSection, error) {
	return note.ReadHeading(b.vaultRoot, path, heading)
}

func (b *NativeBackend) GetBlock(_ context.Context, path, blockID string) (note.Block, error) {
	return note.GetBlock(b.vaultRoot, path, blockID)
}

func (b *NativeBackend) SetBlock(_ context.Context, path, blockID, content string) (note.Block, error) {
	return note.SetBlock(b.vaultRoot, path, blockID, content)
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

func (b *NativeBackend) DailyPath(_ context.Context, at time.Time) (string, error) {
	return note.ResolveDailyPath(b.vaultRoot, at)
}

func (b *NativeBackend) DailyRead(_ context.Context, at time.Time, create bool) (note.Note, error) {
	n, _, err := note.DailyRead(b.vaultRoot, at, create)
	return n, err
}

func (b *NativeBackend) DailyAppend(_ context.Context, at time.Time, content string, inline bool) (note.Note, error) {
	n, _, err := note.DailyAppend(b.vaultRoot, at, content, inline)
	return n, err
}

func (b *NativeBackend) DailyPrepend(_ context.Context, at time.Time, content string, inline bool) (note.Note, error) {
	n, _, err := note.DailyPrepend(b.vaultRoot, at, content, inline)
	return n, err
}

func (b *NativeBackend) ListTemplates(_ context.Context) ([]templates.TemplateInfo, error) {
	return templates.List(b.vaultRoot, b.cfg)
}

func (b *NativeBackend) ReadTemplate(_ context.Context, name, title string, resolve bool) (templates.Template, error) {
	return templates.Read(b.vaultRoot, b.cfg, name, title, resolve, now())
}

func (b *NativeBackend) InsertTemplate(_ context.Context, path, name, title string, resolve bool) (note.Note, error) {
	n, err := note.Get(b.vaultRoot, path)
	if err != nil {
		return note.Note{}, err
	}
	tpl, err := templates.Read(b.vaultRoot, b.cfg, name, title, resolve, now())
	if err != nil {
		return note.Note{}, err
	}
	n.Body += tpl.Content
	return note.Write(b.vaultRoot, n.Path, n, false, now())
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

func (b *NativeBackend) OpenInObsidian(_ context.Context, path string, launch bool) (OpenResult, error) {
	n, err := note.Get(b.vaultRoot, path)
	if err != nil {
		return OpenResult{}, err
	}

	vaultName := filepath.Base(b.vaultRoot)
	uri := fmt.Sprintf("obsidian://open?vault=%s&file=%s", urlQueryEscape(vaultName), urlQueryEscape(strings.TrimSuffix(n.Path, ".md")))
	result := OpenResult{URI: uri}
	if !launch {
		return result, nil
	}

	cmd, args := openCommand(uri)
	if cmd == "" {
		return OpenResult{}, errs.New(errs.ExitGeneric, "launch is not supported on this platform")
	}
	if err := exec.Command(cmd, args...).Run(); err != nil {
		return OpenResult{}, errs.Wrap(errs.ExitGeneric, "failed to launch Obsidian URI", err)
	}
	result.Launched = true
	return result, nil
}

func (b *NativeBackend) SyncStatus(_ context.Context) (SyncStatus, error) {
	status := SyncStatus{
		EffectiveMode: b.mode,
	}

	configPath := filepath.Join(b.vaultRoot, ".obsidian", "sync.json")
	if info, err := os.Stat(configPath); err == nil {
		status.Configured = true
		status.ConfigPath = filepath.ToSlash(filepath.Join(".obsidian", "sync.json"))
		status.LastModified = info.ModTime().UTC().Format(time.RFC3339)
	}

	corePath := filepath.Join(b.vaultRoot, ".obsidian", "core-plugins.json")
	payload, err := os.ReadFile(corePath)
	if err == nil {
		var corePlugins []string
		if parseErr := json.Unmarshal(payload, &corePlugins); parseErr == nil {
			for _, id := range corePlugins {
				if strings.TrimSpace(id) == "sync" {
					status.Enabled = true
					break
				}
			}
		}
	}
	return status, nil
}

func (b *NativeBackend) ListTasks(_ context.Context, opts tasks.ListOptions) ([]tasks.Task, error) {
	return tasks.List(b.vaultRoot, opts)
}

func (b *NativeBackend) GetTask(_ context.Context, ref tasks.Ref) (tasks.Task, error) {
	return tasks.Get(b.vaultRoot, ref)
}

func (b *NativeBackend) UpdateTask(_ context.Context, ref tasks.Ref, input tasks.UpdateInput) (tasks.Task, error) {
	return tasks.Update(b.vaultRoot, ref, input)
}

func (b *NativeBackend) ListPlugins(_ context.Context, filter string, enabledOnly bool) ([]PluginInfo, error) {
	filter = strings.TrimSpace(strings.ToLower(filter))
	if filter != "" && filter != "core" && filter != "community" {
		return nil, errs.New(errs.ExitValidation, "filter must be core or community")
	}

	coreEnabled, _ := b.loadPluginIDs(filepath.Join(".obsidian", "core-plugins.json"))
	communityEnabled, _ := b.loadPluginIDs(filepath.Join(".obsidian", "community-plugins.json"))
	items := []PluginInfo{}

	if filter == "" || filter == "core" {
		for _, id := range coreEnabled {
			items = append(items, PluginInfo{
				ID:      id,
				Name:    id,
				Type:    "core",
				Enabled: true,
			})
		}
	}

	if filter == "" || filter == "community" {
		communityDir := filepath.Join(b.vaultRoot, ".obsidian", "plugins")
		entries, err := os.ReadDir(communityDir)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				id := entry.Name()
				info := PluginInfo{
					ID:      id,
					Type:    "community",
					Enabled: contains(communityEnabled, id),
				}
				manifestPath := filepath.Join(communityDir, id, "manifest.json")
				payload, readErr := os.ReadFile(manifestPath)
				if readErr == nil {
					var manifest struct {
						Name    string `json:"name"`
						Version string `json:"version"`
					}
					if jsonErr := json.Unmarshal(payload, &manifest); jsonErr == nil {
						info.Name = manifest.Name
						info.Version = manifest.Version
					}
				}
				if info.Name == "" {
					info.Name = id
				}
				items = append(items, info)
			}
		}
	}

	if enabledOnly {
		filtered := make([]PluginInfo, 0, len(items))
		for _, item := range items {
			if item.Enabled {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Type == items[j].Type {
			return items[i].ID < items[j].ID
		}
		return items[i].Type < items[j].Type
	})
	return items, nil
}

func (b *NativeBackend) ListCommandIDs(_ context.Context, filter string) ([]string, error) {
	hotkeysPath := filepath.Join(b.vaultRoot, ".obsidian", "hotkeys.json")
	payload, err := os.ReadFile(hotkeysPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var hotkeys map[string]any
	if err := json.Unmarshal(payload, &hotkeys); err != nil {
		return nil, err
	}

	out := make([]string, 0, len(hotkeys))
	prefix := strings.TrimSpace(filter)
	for commandID := range hotkeys {
		if prefix != "" && !strings.HasPrefix(commandID, prefix) {
			continue
		}
		out = append(out, commandID)
	}
	sort.Strings(out)
	return out, nil
}

func (b *NativeBackend) ExecuteCommand(_ context.Context, _ string) error {
	return errs.New(errs.ExitGeneric, "command execution requires Obsidian app/API mode")
}

func (b *NativeBackend) loadPluginIDs(relPath string) ([]string, error) {
	payload, err := os.ReadFile(filepath.Join(b.vaultRoot, relPath))
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	var ids []string
	if err := json.Unmarshal(payload, &ids); err != nil {
		return nil, err
	}
	return ids, nil
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

func contains(values []string, candidate string) bool {
	for _, value := range values {
		if value == candidate {
			return true
		}
	}
	return false
}

func urlQueryEscape(value string) string {
	replacer := strings.NewReplacer(
		"%", "%25",
		" ", "%20",
		"\"", "%22",
		"#", "%23",
		"&", "%26",
		"+", "%2B",
		",", "%2C",
		"/", "%2F",
		":", "%3A",
		";", "%3B",
		"=", "%3D",
		"?", "%3F",
		"@", "%40",
	)
	return replacer.Replace(value)
}

func openCommand(uri string) (string, []string) {
	switch runtime.GOOS {
	case "darwin":
		return "open", []string{uri}
	case "linux":
		return "xdg-open", []string{uri}
	case "windows":
		return "rundll32", []string{"url.dll,FileProtocolHandler", uri}
	default:
		return "", nil
	}
}
