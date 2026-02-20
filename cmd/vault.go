package cmd

import "github.com/spf13/cobra"

func newVaultCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vault",
		Short: "Vault management commands",
	}
	cmd.AddCommand(newVaultInitCmd())
	cmd.AddCommand(newVaultStatusCmd())
	return cmd
}
