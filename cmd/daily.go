package cmd

import "github.com/spf13/cobra"

func newDailyCmd() *cobra.Command {
	var date string
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "daily",
		Short: "Open or create today's daily note",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			at, err := parseDateFlag(date)
			if err != nil {
				return err
			}
			if dryRun {
				path, err := rt.Backend.DailyPath(rt.Context, at)
				if err != nil {
					return err
				}
				if rt.Printer.JSON {
					return rt.Printer.PrintJSON(map[string]any{
						"dry_run": true,
						"action":  "daily.read-or-create",
						"path":    path,
						"date":    at.Format("2006-01-02"),
					})
				}
				rt.Printer.Println("dry-run: would read/create " + path)
				return nil
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
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview operation without writing files")
	cmd.AddCommand(newDailyPathCmd())
	cmd.AddCommand(newDailyReadCmd())
	cmd.AddCommand(newDailyAppendCmd())
	cmd.AddCommand(newDailyPrependCmd())
	return cmd
}

func newDailyPathCmd() *cobra.Command {
	var date string
	cmd := &cobra.Command{
		Use:   "path",
		Short: "Get daily note path",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			at, err := parseDateFlag(date)
			if err != nil {
				return err
			}
			path, err := rt.Backend.DailyPath(rt.Context, at)
			if err != nil {
				return err
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(map[string]any{"path": path})
			}
			rt.Printer.Println(path)
			return nil
		},
	}
	cmd.Flags().StringVar(&date, "date", "", "Date (YYYY-MM-DD), default today")
	return cmd
}

func newDailyReadCmd() *cobra.Command {
	var date string
	cmd := &cobra.Command{
		Use:   "read",
		Short: "Read daily note",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			at, err := parseDateFlag(date)
			if err != nil {
				return err
			}
			n, err := rt.Backend.DailyRead(rt.Context, at, false)
			if err != nil {
				return err
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(n)
			}
			rt.Printer.Println(n.Raw)
			return nil
		},
	}
	cmd.Flags().StringVar(&date, "date", "", "Date (YYYY-MM-DD), default today")
	return cmd
}

func newDailyAppendCmd() *cobra.Command {
	var date string
	var inline bool
	var dryRun bool
	var ifHash string
	cmd := &cobra.Command{
		Use:   "append <content>",
		Short: "Append content to daily note",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			at, err := parseDateFlag(date)
			if err != nil {
				return err
			}
			path, err := rt.Backend.DailyPath(rt.Context, at)
			if err != nil {
				return err
			}
			if err := verifyHashPrecondition(rt, path, ifHash); err != nil {
				return err
			}
			if dryRun {
				if rt.Printer.JSON {
					return rt.Printer.PrintJSON(map[string]any{
						"dry_run": true,
						"action":  "daily.append",
						"path":    path,
						"content": args[0],
						"inline":  inline,
					})
				}
				rt.Printer.Println("dry-run: would append to " + path)
				return nil
			}
			n, err := rt.Backend.DailyAppend(rt.Context, at, args[0], inline)
			if err != nil {
				return err
			}
			if err := enforceWriteGuards(rt, n); err != nil {
				return err
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(n)
			}
			rt.Printer.Println("updated: " + n.Path)
			return nil
		},
	}
	cmd.Flags().StringVar(&date, "date", "", "Date (YYYY-MM-DD), default today")
	cmd.Flags().BoolVar(&inline, "inline", false, "Append without newline")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview operation without writing files")
	cmd.Flags().StringVar(&ifHash, "if-hash", "", "Require current note SHA256 hash before writing")
	return cmd
}

func newDailyPrependCmd() *cobra.Command {
	var date string
	var inline bool
	var dryRun bool
	var ifHash string
	cmd := &cobra.Command{
		Use:   "prepend <content>",
		Short: "Prepend content to daily note",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			at, err := parseDateFlag(date)
			if err != nil {
				return err
			}
			path, err := rt.Backend.DailyPath(rt.Context, at)
			if err != nil {
				return err
			}
			if err := verifyHashPrecondition(rt, path, ifHash); err != nil {
				return err
			}
			if dryRun {
				if rt.Printer.JSON {
					return rt.Printer.PrintJSON(map[string]any{
						"dry_run": true,
						"action":  "daily.prepend",
						"path":    path,
						"content": args[0],
						"inline":  inline,
					})
				}
				rt.Printer.Println("dry-run: would prepend to " + path)
				return nil
			}
			n, err := rt.Backend.DailyPrepend(rt.Context, at, args[0], inline)
			if err != nil {
				return err
			}
			if err := enforceWriteGuards(rt, n); err != nil {
				return err
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(n)
			}
			rt.Printer.Println("updated: " + n.Path)
			return nil
		},
	}
	cmd.Flags().StringVar(&date, "date", "", "Date (YYYY-MM-DD), default today")
	cmd.Flags().BoolVar(&inline, "inline", false, "Prepend without newline")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview operation without writing files")
	cmd.Flags().StringVar(&ifHash, "if-hash", "", "Require current note SHA256 hash before writing")
	return cmd
}
