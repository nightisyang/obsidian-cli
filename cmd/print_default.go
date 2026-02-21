package cmd

import "github.com/spf13/cobra"

func newPrintDefaultCmd() *cobra.Command {
	var pathOnly bool
	cmd := &cobra.Command{
		Use:   "print-default",
		Short: "Print resolved vault path and config source",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			if pathOnly {
				rt.Printer.Println(rt.VaultRoot)
				return nil
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(map[string]any{
					"vault_path":     rt.VaultRoot,
					"config_path":    rt.ConfigPath,
					"config_source":  rt.ConfigSource,
					"requested_mode": rt.RequestedMode,
					"effective_mode": rt.EffectiveMode,
				})
			}
			rt.Printer.Println("vault path: " + rt.VaultRoot)
			if rt.ConfigPath != "" {
				rt.Printer.Println("config path: " + rt.ConfigPath)
			}
			rt.Printer.Println("config source: " + rt.ConfigSource)
			rt.Printer.Println("effective mode: " + rt.EffectiveMode)
			return nil
		},
	}
	cmd.Flags().BoolVar(&pathOnly, "path-only", false, "Print only vault path")
	return cmd
}
