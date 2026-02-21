package cmd

import (
	"github.com/nightisyang/obsidian-cli/internal/note"
	"github.com/spf13/cobra"
)

func newNoteMoveCmd() *cobra.Command {
	var updateLinks bool
	var dryRun bool
	var ifHash string

	cmd := &cobra.Command{
		Use:   "move <src> <dst>",
		Short: "Move a note and optionally rewrite links",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			if err := verifyHashPrecondition(rt, args[0], ifHash); err != nil {
				return err
			}
			n, err := rt.Backend.MoveNote(rt.Context, args[0], args[1], note.MoveOptions{UpdateLinks: updateLinks, DryRun: dryRun})
			if err != nil {
				return err
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(map[string]any{"note": n, "dry_run": dryRun, "update_links": updateLinks})
			}
			rt.Printer.Println("moved: " + n.Path)
			return nil
		},
	}
	cmd.Flags().BoolVar(&updateLinks, "update-links", true, "Rewrite wikilinks that point to the source note")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would happen without moving")
	cmd.Flags().StringVar(&ifHash, "if-hash", "", "Require source note SHA256 hash before writing")
	return cmd
}
