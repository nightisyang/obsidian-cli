package cmd

import "github.com/spf13/cobra"

func newNoteAppendCmd() *cobra.Command {
	var dryRun bool
	var ifHash string
	cmd := &cobra.Command{
		Use:   "append <path> <content>",
		Short: "Append content to a note",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			if err := verifyHashPrecondition(rt, args[0], ifHash); err != nil {
				return err
			}
			if dryRun {
				current, err := rt.Backend.GetNote(rt.Context, args[0])
				if err != nil {
					return err
				}
				nextBody := current.Body
				if nextBody != "" && nextBody[len(nextBody)-1] != '\n' {
					nextBody += "\n"
				}
				nextBody += args[1]
				if rt.Printer.JSON {
					return rt.Printer.PrintJSON(map[string]any{
						"dry_run":     true,
						"action":      "note.append",
						"path":        current.Path,
						"hash_before": hashString(current.Raw),
						"hash_after":  hashString(nextBody),
					})
				}
				rt.Printer.Println("dry-run: would append to " + current.Path)
				return nil
			}
			n, err := rt.Backend.AppendNote(rt.Context, args[0], args[1])
			if err != nil {
				return err
			}
			if err := enforceWriteGuards(rt, n); err != nil {
				return err
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(n)
			}
			rt.Printer.Println("updated: " + n.Path)
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview operation without writing files")
	cmd.Flags().StringVar(&ifHash, "if-hash", "", "Require current note SHA256 hash before writing")
	return cmd
}
