package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync operations",
	}
	cmd.AddCommand(newSyncStatusCmd())
	return cmd
}

func newSyncStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show sync status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			status, err := rt.Backend.SyncStatus(rt.Context)
			if err != nil {
				return err
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(status)
			}
			rt.Printer.Println(fmt.Sprintf("configured: %t", status.Configured))
			rt.Printer.Println(fmt.Sprintf("enabled: %t", status.Enabled))
			if status.ConfigPath != "" {
				rt.Printer.Println("config: " + status.ConfigPath)
			}
			if status.LastModified != "" {
				rt.Printer.Println("updated: " + status.LastModified)
			}
			return nil
		},
	}
	return cmd
}
