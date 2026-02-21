package cmd

import "github.com/spf13/cobra"

func newBlockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block",
		Short: "Block reference operations",
	}
	cmd.AddCommand(newBlockGetCmd())
	cmd.AddCommand(newBlockSetCmd())
	return cmd
}

func newBlockGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <path> <block-id>",
		Short: "Read content by block ID",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			block, err := rt.Backend.GetBlock(rt.Context, args[0], args[1])
			if err != nil {
				return err
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(block)
			}
			rt.Printer.Println(block.Content)
			return nil
		},
	}
	return cmd
}

func newBlockSetCmd() *cobra.Command {
	var dryRun bool
	var ifHash string
	cmd := &cobra.Command{
		Use:   "set <path> <block-id> <content>",
		Short: "Update content by block ID",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			if err := verifyHashPrecondition(rt, args[0], ifHash); err != nil {
				return err
			}
			if dryRun {
				existing, err := rt.Backend.GetBlock(rt.Context, args[0], args[1])
				if err != nil {
					return err
				}
				if rt.Printer.JSON {
					return rt.Printer.PrintJSON(map[string]any{
						"dry_run":     true,
						"action":      "block.set",
						"path":        existing.Path,
						"block_id":    existing.BlockID,
						"old_content": existing.Content,
						"new_content": args[2],
					})
				}
				rt.Printer.Println("dry-run: would update block " + existing.BlockID + " in " + existing.Path)
				return nil
			}
			block, err := rt.Backend.SetBlock(rt.Context, args[0], args[1], args[2])
			if err != nil {
				return err
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(block)
			}
			rt.Printer.Println("updated: " + block.Path)
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview operation without writing files")
	cmd.Flags().StringVar(&ifHash, "if-hash", "", "Require current note SHA256 hash before writing")
	return cmd
}
