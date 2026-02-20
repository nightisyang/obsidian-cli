package cmd

import "github.com/spf13/cobra"

func newTemplatesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "templates",
		Short: "List templates",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			templates, err := rt.Backend.ListTemplates(rt.Context)
			if err != nil {
				return err
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(templates)
			}
			for _, item := range templates {
				rt.Printer.Println(item.Name)
			}
			return nil
		},
	}
	return cmd
}

func newTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Template operations",
	}
	cmd.AddCommand(newTemplateReadCmd())
	cmd.AddCommand(newTemplateInsertCmd())
	return cmd
}

func newTemplateReadCmd() *cobra.Command {
	var title string
	var resolve bool
	cmd := &cobra.Command{
		Use:   "read <name>",
		Short: "Read template content",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			tpl, err := rt.Backend.ReadTemplate(rt.Context, args[0], title, resolve)
			if err != nil {
				return err
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(tpl)
			}
			rt.Printer.Println(tpl.Content)
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "Title for variable resolution")
	cmd.Flags().BoolVar(&resolve, "resolve", false, "Resolve {{date}}, {{time}}, and {{title}}")
	return cmd
}

func newTemplateInsertCmd() *cobra.Command {
	var title string
	var resolve bool
	cmd := &cobra.Command{
		Use:   "insert <path> <name>",
		Short: "Insert template into a note",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			n, err := rt.Backend.InsertTemplate(rt.Context, args[0], args[1], title, resolve)
			if err != nil {
				return err
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(n)
			}
			rt.Printer.Println("updated: " + n.Path)
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "Title for variable resolution")
	cmd.Flags().BoolVar(&resolve, "resolve", true, "Resolve {{date}}, {{time}}, and {{title}}")
	return cmd
}
