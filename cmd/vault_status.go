package cmd

import (
	"fmt"

	"github.com/nightisyang/obsidian-cli/internal/vault"
	"github.com/spf13/cobra"
)

func newVaultStatusCmd() *cobra.Command {
	var reindex bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show vault status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			_ = reindex
			runtime, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			status, err := vault.ComputeStatus(runtime.VaultRoot, runtime.ConfigPath, runtime.ConfigSource, runtime.EffectiveMode)
			if err != nil {
				return err
			}
			if runtime.Printer.JSON {
				return runtime.Printer.PrintJSON(status)
			}
			if !runtime.Printer.Quiet {
				runtime.Printer.Println("Vault status")
			}
			runtime.Printer.Printf("root: %s\n", status.Root)
			runtime.Printer.Printf("notes: %d\n", status.NoteCount)
			runtime.Printer.Printf("tags: %d\n", status.TagCount)
			runtime.Printer.Printf("mode: %s\n", status.EffectiveMode)
			if status.ConfigPath != "" {
				runtime.Printer.Printf("config: %s (%s)\n", status.ConfigPath, status.ConfigSource)
			} else {
				runtime.Printer.Printf("config: %s\n", status.ConfigSource)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&reindex, "reindex", false, "Rebuild indexes before status")
	cmd.SetHelpTemplate(fmt.Sprintf("%s", cmd.HelpTemplate()))
	return cmd
}
