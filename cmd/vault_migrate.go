package cmd

import (
	"fmt"

	"github.com/nightisyang/obsidian-cli/internal/vault"
	"github.com/spf13/cobra"
)

func newVaultMigrateCmd() *cobra.Command {
	var dryRun bool
	var defaultKind string

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate notes without frontmatter",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}

			result, err := vault.MigrateFrontmatter(rt.VaultRoot, vault.MigrateOptions{
				DryRun: dryRun,
				Kind:   defaultKind,
			})
			if err != nil {
				return err
			}

			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(result)
			}
			if dryRun {
				for _, file := range result.Files {
					if file.Action == "would_migrate" {
						rt.Printer.Println("would migrate: " + file.Path)
					}
				}
			}
			rt.Printer.Println(fmt.Sprintf("Migrated %d notes, %d already had frontmatter", result.Migrated, result.Skipped))
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print what would change without writing files")
	cmd.Flags().StringVar(&defaultKind, "kind", "note", "Default kind for migrated notes")
	return cmd
}
