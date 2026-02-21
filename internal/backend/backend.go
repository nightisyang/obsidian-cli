package backend

import (
	"context"
	"time"

	"github.com/nightisyang/obsidian-cli/internal/index"
	"github.com/nightisyang/obsidian-cli/internal/note"
	"github.com/nightisyang/obsidian-cli/internal/search"
	"github.com/nightisyang/obsidian-cli/internal/tasks"
	"github.com/nightisyang/obsidian-cli/internal/templates"
	"github.com/nightisyang/obsidian-cli/internal/vault"
)

type OpenResult struct {
	URI      string `json:"uri"`
	Launched bool   `json:"launched"`
}

type SyncStatus struct {
	Configured    bool   `json:"configured"`
	Enabled       bool   `json:"enabled"`
	ConfigPath    string `json:"config_path,omitempty"`
	LastModified  string `json:"last_modified,omitempty"`
	EffectiveMode string `json:"effective_mode"`
}

type PluginInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
}

type Backend interface {
	VaultStatus(ctx context.Context) (vault.Status, error)
	CreateNote(ctx context.Context, in note.CreateInput) (note.Note, error)
	GetNote(ctx context.Context, path string) (note.Note, error)
	GetHeading(ctx context.Context, path, heading string) (note.HeadingSection, error)
	GetBlock(ctx context.Context, path, blockID string) (note.Block, error)
	SetBlock(ctx context.Context, path, blockID, content string) (note.Block, error)
	AppendNote(ctx context.Context, path, content string) (note.Note, error)
	PrependNote(ctx context.Context, path, content string) (note.Note, error)
	DeleteNote(ctx context.Context, path string) error
	ListNotes(ctx context.Context, dir string, opts note.ListOptions) ([]note.Note, error)
	MoveNote(ctx context.Context, src, dst string, opts note.MoveOptions) (note.Note, error)
	DailyPath(ctx context.Context, at time.Time) (string, error)
	DailyRead(ctx context.Context, at time.Time, create bool) (note.Note, error)
	DailyAppend(ctx context.Context, at time.Time, content string, inline bool) (note.Note, error)
	DailyPrepend(ctx context.Context, at time.Time, content string, inline bool) (note.Note, error)
	ListTemplates(ctx context.Context) ([]templates.TemplateInfo, error)
	ReadTemplate(ctx context.Context, name, title string, resolve bool) (templates.Template, error)
	InsertTemplate(ctx context.Context, path, name, title string, resolve bool) (note.Note, error)
	Search(ctx context.Context, q search.Query) ([]search.SearchResult, error)
	ListTags(ctx context.Context, opts index.TagListOptions) ([]index.TagCount, error)
	SearchTag(ctx context.Context, tag string, limit int) ([]search.SearchResult, error)
	OutgoingLinks(ctx context.Context, path string) ([]string, error)
	Backlinks(ctx context.Context, path string, rebuild bool) ([]string, error)
	PropGet(ctx context.Context, path, key string) (any, error)
	PropSet(ctx context.Context, path, key string, value any) (note.Note, error)
	PropDelete(ctx context.Context, path, key string) (note.Note, error)
	PropList(ctx context.Context, path string) (map[string]any, error)
	OpenInObsidian(ctx context.Context, path string, launch bool) (OpenResult, error)
	SyncStatus(ctx context.Context) (SyncStatus, error)
	ListTasks(ctx context.Context, opts tasks.ListOptions) ([]tasks.Task, error)
	GetTask(ctx context.Context, ref tasks.Ref) (tasks.Task, error)
	UpdateTask(ctx context.Context, ref tasks.Ref, input tasks.UpdateInput) (tasks.Task, error)
	ListPlugins(ctx context.Context, filter string, enabledOnly bool) ([]PluginInfo, error)
	ListCommandIDs(ctx context.Context, filter string) ([]string, error)
	ExecuteCommand(ctx context.Context, id string) error
}
