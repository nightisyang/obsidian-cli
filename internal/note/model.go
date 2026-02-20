package note

import "github.com/nightisyang/obsidian-cli/internal/frontmatter"

type Frontmatter = frontmatter.Data

type Note struct {
	Path        string      `json:"path"`
	Title       string      `json:"title"`
	Frontmatter Frontmatter `json:"frontmatter"`
	Body        string      `json:"body"`
	Raw         string      `json:"raw,omitempty"`
	Outgoing    []string    `json:"outgoing_links,omitempty"`
	InlineTags  []string    `json:"inline_tags,omitempty"`
}

type CreateInput struct {
	Title    string
	Dir      string
	Template string
	Tags     []string
	Kind     string
	Topic    string
	Status   string
}

type ListOptions struct {
	Recursive bool
	Limit     int
	Sort      string
}

type MoveOptions struct {
	UpdateLinks bool
	DryRun      bool
}
