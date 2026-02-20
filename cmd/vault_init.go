package cmd

import (
	"path/filepath"

	"github.com/nightisyang/obsidian-cli/internal/errs"
	"github.com/nightisyang/obsidian-cli/internal/vault"
	"github.com/spf13/cobra"
)

func newVaultInitCmd() *cobra.Command {
	var force bool
	var name string
	var modeDefault string

	cmd := &cobra.Command{
		Use:   "init [path]",
		Short: "Initialize a vault config",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := "."
			if len(args) == 1 {
				target = args[0]
			}
			root, err := filepath.Abs(target)
			if err != nil {
				return errs.Wrap(errs.ExitValidation, "invalid target path", err)
			}

			cfg := vault.DefaultConfig()
			cfg.VaultPath = "."
			if modeDefault != "" {
				cfg.ModeDefault = modeDefault
			}

			configPath := filepath.Join(root, ".obsidian-cli.yaml")
			if err := vault.WriteConfig(configPath, cfg, force); err != nil {
				return err
			}

			runtime, _ := getRuntime(cmd)
			data := map[string]any{
				"path":         root,
				"config_path":  configPath,
				"name":         name,
				"mode_default": cfg.ModeDefault,
			}
			if runtime != nil && runtime.Printer.JSON {
				return runtime.Printer.PrintJSON(data)
			}
			if runtime != nil {
				runtime.Printer.Println("Initialized vault config: " + configPath)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing config")
	cmd.Flags().StringVar(&name, "name", "", "Vault display name")
	cmd.Flags().StringVar(&modeDefault, "mode-default", "", "Default mode")
	return cmd
}
