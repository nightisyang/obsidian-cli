package cmd

import "github.com/spf13/cobra"

func newPropDeleteCmd() *cobra.Command {
	var dryRun bool
	var ifHash string
	cmd := &cobra.Command{
		Use:   "delete <path> <key>",
		Short: "Delete a frontmatter property",
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
				if rt.Printer.JSON {
					return rt.Printer.PrintJSON(map[string]any{
						"dry_run": true,
						"path":    args[0],
						"key":     args[1],
						"action":  "prop.delete",
					})
				}
				rt.Printer.Println("dry-run: would delete property " + args[1] + " from " + args[0])
				return nil
			}
			n, err := rt.Backend.PropDelete(rt.Context, args[0], args[1])
			if err != nil {
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
