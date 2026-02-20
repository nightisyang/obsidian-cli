package cmd

import "github.com/spf13/cobra"

func newTagCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "tag", Short: "Tag operations"}
	cmd.AddCommand(newTagListCmd())
	cmd.AddCommand(newTagSearchCmd())
	return cmd
}
