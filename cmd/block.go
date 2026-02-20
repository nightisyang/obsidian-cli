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
	cmd := &cobra.Command{
		Use:   "set <path> <block-id> <content>",
		Short: "Update content by block ID",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
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
	return cmd
}
