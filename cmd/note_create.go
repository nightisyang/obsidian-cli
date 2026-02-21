package cmd

import (
	"regexp"
	"strings"

	"github.com/nightisyang/obsidian-cli/internal/note"
	"github.com/spf13/cobra"
)

func newNoteCreateCmd() *cobra.Command {
	var template string
	var content string
	var tags []string
	var dir string
	var kind string
	var topic string
	var status string
	var openOnly bool
	var dryRun bool
	var ifMissing bool

	cmd := &cobra.Command{
		Use:   "create <title>",
		Short: "Create a new note",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			predictedPath := predictCreatePath(args[0], dir)
			if ifMissing {
				if _, existingErr := rt.Backend.GetNote(rt.Context, predictedPath); existingErr == nil {
					if rt.Printer.JSON {
						return rt.Printer.PrintJSON(map[string]any{
							"skipped": true,
							"reason":  "note exists",
							"path":    predictedPath,
						})
					}
					rt.Printer.Println("skipped (exists): " + predictedPath)
					return nil
				}
			}
			if dryRun {
				if rt.Printer.JSON {
					return rt.Printer.PrintJSON(map[string]any{
						"dry_run":        true,
						"action":         "note.create",
						"title":          args[0],
						"predicted_path": predictedPath,
						"template":       template,
						"tags":           tags,
						"kind":           kind,
						"topic":          topic,
						"status":         status,
					})
				}
				rt.Printer.Println("dry-run: would create " + predictedPath)
				return nil
			}
			created, err := rt.Backend.CreateNote(rt.Context, note.CreateInput{
				Title:    args[0],
				Template: template,
				Content:  content,
				Tags:     tags,
				Dir:      dir,
				Kind:     kind,
				Topic:    topic,
				Status:   status,
			})
			if err != nil {
				return err
			}
			if openOnly {
				if rt.Printer.JSON {
					return rt.Printer.PrintJSON(map[string]any{"path": created.Path})
				}
				rt.Printer.Println(created.Path)
				return nil
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(created)
			}
			rt.Printer.Println("created: " + created.Path)
			return nil
		},
	}

	cmd.Flags().StringVar(&template, "template", "", "Template name")
	cmd.Flags().StringVar(&content, "content", "", "Initial note body content")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "Add tag (repeatable)")
	cmd.Flags().StringVar(&dir, "dir", "", "Destination directory")
	cmd.Flags().StringVar(&kind, "kind", "", "Note kind")
	cmd.Flags().StringVar(&topic, "topic", "", "Topic")
	cmd.Flags().StringVar(&status, "status", "", "Status")
	cmd.Flags().BoolVar(&openOnly, "open", false, "Print resulting path only")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview operation without writing files")
	cmd.Flags().BoolVar(&ifMissing, "if-missing", false, "No-op successfully when target note already exists")
	return cmd
}

func predictCreatePath(title, dir string) string {
	trimmedDir := strings.Trim(strings.TrimSpace(dir), "/")
	slug := slugForCreate(title)
	rel := slug + ".md"
	if trimmedDir == "" {
		return rel
	}
	return trimmedDir + "/" + rel
}

func slugForCreate(input string) string {
	lower := strings.ToLower(strings.TrimSpace(input))
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug := re.ReplaceAllString(lower, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		return "untitled"
	}
	return slug
}
