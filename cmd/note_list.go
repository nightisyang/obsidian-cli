package cmd

import (
	"fmt"

	"github.com/nightisyang/obsidian-cli/internal/note"
	"github.com/spf13/cobra"
)

func newNoteListCmd() *cobra.Command {
	var recursive bool
	var limit int
	var sortBy string

	cmd := &cobra.Command{
		Use:   "list [dir]",
		Short: "List notes",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			dir := ""
			if len(args) == 1 {
				dir = args[0]
			}
			notes, err := rt.Backend.ListNotes(rt.Context, dir, note.ListOptions{Recursive: recursive, Limit: limit, Sort: sortBy})
			if err != nil {
				return err
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(notes)
			}
			for _, n := range notes {
				rt.Printer.Println(fmt.Sprintf("%s\t%s", n.Path, n.Title))
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&recursive, "recursive", true, "Recurse into subdirectories")
	cmd.Flags().IntVar(&limit, "limit", 0, "Limit number of notes")
	cmd.Flags().StringVar(&sortBy, "sort", "name", "Sort by: name|updated|created")
	return cmd
}
