package backend

import (
	"context"

	"github.com/nightisyang/obsidian-cli/internal/index"
	"github.com/nightisyang/obsidian-cli/internal/note"
	"github.com/nightisyang/obsidian-cli/internal/search"
	"github.com/nightisyang/obsidian-cli/internal/vault"
)

type Backend interface {
	VaultStatus(ctx context.Context) (vault.Status, error)
	CreateNote(ctx context.Context, in note.CreateInput) (note.Note, error)
	GetNote(ctx context.Context, path string) (note.Note, error)
	AppendNote(ctx context.Context, path, content string) (note.Note, error)
	PrependNote(ctx context.Context, path, content string) (note.Note, error)
	DeleteNote(ctx context.Context, path string) error
	ListNotes(ctx context.Context, dir string, opts note.ListOptions) ([]note.Note, error)
	MoveNote(ctx context.Context, src, dst string, opts note.MoveOptions) (note.Note, error)
	Search(ctx context.Context, q search.Query) ([]search.SearchResult, error)
	ListTags(ctx context.Context, opts index.TagListOptions) ([]index.TagCount, error)
	SearchTag(ctx context.Context, tag string, limit int) ([]search.SearchResult, error)
	OutgoingLinks(ctx context.Context, path string) ([]string, error)
	Backlinks(ctx context.Context, path string, rebuild bool) ([]string, error)
	PropGet(ctx context.Context, path, key string) (any, error)
	PropSet(ctx context.Context, path, key string, value any) (note.Note, error)
	PropList(ctx context.Context, path string) (map[string]any, error)
}
