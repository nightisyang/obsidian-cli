package cmd

import "github.com/spf13/cobra"

func newPropCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "prop", Short: "Property operations"}
	cmd.AddCommand(newPropGetCmd())
	cmd.AddCommand(newPropSetCmd())
	cmd.AddCommand(newPropDeleteCmd())
	cmd.AddCommand(newPropListCmd())
	return cmd
}
