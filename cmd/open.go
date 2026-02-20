package cmd

import "github.com/spf13/cobra"

func newOpenCmd() *cobra.Command {
	var launch bool
	cmd := &cobra.Command{
		Use:   "open <path>",
		Short: "Open a note in the Obsidian app via obsidian:// URI",
		Args:  cobra.ExactArgs(1),
		Long:  "Open a note in Obsidian. Use --launch to execute the URI via the OS; this requires Obsidian app integration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			result, err := rt.Backend.OpenInObsidian(rt.Context, args[0], launch)
			if err != nil {
				return err
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(result)
			}
			rt.Printer.Println(result.URI)
			return nil
		},
	}
	cmd.Flags().BoolVar(&launch, "launch", false, "Launch URI with OS handler")
	return cmd
}
