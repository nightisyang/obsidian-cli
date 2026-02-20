package cmd

import "github.com/spf13/cobra"

func newLinksCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "links", Short: "Link index operations"}
	cmd.AddCommand(newLinksListCmd())
	cmd.AddCommand(newLinksBacklinksCmd())
	return cmd
}
