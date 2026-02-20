package cmd

import "github.com/spf13/cobra"

func newNoteDailyCmd() *cobra.Command {
	var date string
	cmd := &cobra.Command{
		Use:   "daily",
		Short: "Open or create daily note (alias of `daily`)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			at, err := parseDateFlag(date)
			if err != nil {
				return err
			}
			n, err := rt.Backend.DailyRead(rt.Context, at, true)
			if err != nil {
				return err
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(n)
			}
			rt.Printer.Println(n.Path)
			return nil
		},
	}
	cmd.Flags().StringVar(&date, "date", "", "Date (YYYY-MM-DD), default today")
	return cmd
}
