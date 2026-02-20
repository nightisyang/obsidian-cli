package cmd

import (
	"github.com/nightisyang/obsidian-cli/internal/note"
	"github.com/spf13/cobra"
)

func newNoteCreateCmd() *cobra.Command {
	var template string
	var tags []string
	var dir string
	var kind string
	var topic string
	var status string
	var openOnly bool

	cmd := &cobra.Command{
		Use:   "create <title>",
		Short: "Create a new note",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			created, err := rt.Backend.CreateNote(rt.Context, note.CreateInput{
				Title:    args[0],
				Template: template,
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
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "Add tag (repeatable)")
	cmd.Flags().StringVar(&dir, "dir", "", "Destination directory")
	cmd.Flags().StringVar(&kind, "kind", "", "Note kind")
	cmd.Flags().StringVar(&topic, "topic", "", "Topic")
	cmd.Flags().StringVar(&status, "status", "", "Status")
	cmd.Flags().BoolVar(&openOnly, "open", false, "Print resulting path only")
	return cmd
}
