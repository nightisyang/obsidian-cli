package cmd

import "github.com/spf13/cobra"

func newNoteDeleteCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "delete <path>",
		Short: "Delete a note",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = force
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			if err := rt.Backend.DeleteNote(rt.Context, args[0]); err != nil {
				return err
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(map[string]any{"deleted": args[0]})
			}
			rt.Printer.Println("deleted: " + args[0])
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation")
	return cmd
}
