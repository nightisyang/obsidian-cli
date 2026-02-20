package cmd

import (
	"fmt"
	"strings"

	"github.com/nightisyang/obsidian-cli/internal/app"
	"github.com/nightisyang/obsidian-cli/internal/errs"
	"github.com/nightisyang/obsidian-cli/internal/tasks"
	"github.com/spf13/cobra"
)

func newTasksCmd() *cobra.Command {
	var path string
	var status string
	var done bool
	var todo bool
	var total bool
	var daily bool
	var date string
	var limit int

	cmd := &cobra.Command{
		Use:   "tasks",
		Short: "List checkbox tasks",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if done && todo {
				return errs.New(errs.ExitValidation, "--done and --todo are mutually exclusive")
			}
			if strings.TrimSpace(status) != "" && (done || todo) {
				return errs.New(errs.ExitValidation, "--status cannot be combined with --done/--todo")
			}
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}

			resolvedPath := path
			if daily {
				at, err := parseDateFlag(date)
				if err != nil {
					return err
				}
				resolvedPath, err = rt.Backend.DailyPath(rt.Context, at)
				if err != nil {
					return err
				}
			}

			list, err := rt.Backend.ListTasks(rt.Context, tasks.ListOptions{
				Path:   resolvedPath,
				Status: status,
				Done:   done,
				Todo:   todo,
				Limit:  limit,
			})
			if err != nil {
				return err
			}

			if total {
				if rt.Printer.JSON {
					return rt.Printer.PrintJSON(map[string]any{"total": len(list)})
				}
				rt.Printer.Println(fmt.Sprintf("%d", len(list)))
				return nil
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(list)
			}
			for _, task := range list {
				rt.Printer.Println(fmt.Sprintf("%s:%d [%s] %s", task.Path, task.Line, task.Status, task.Text))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "Filter by note path")
	cmd.Flags().StringVar(&status, "status", "", "Filter by single status character")
	cmd.Flags().BoolVar(&done, "done", false, "Only completed tasks ([x])")
	cmd.Flags().BoolVar(&todo, "todo", false, "Only incomplete tasks ([ ])")
	cmd.Flags().BoolVar(&total, "total", false, "Return count only")
	cmd.Flags().BoolVar(&daily, "daily", false, "Use daily note path")
	cmd.Flags().StringVar(&date, "date", "", "Daily note date (YYYY-MM-DD)")
	cmd.Flags().IntVar(&limit, "limit", 0, "Limit result count")
	return cmd
}

func newTaskCmd() *cobra.Command {
	var refRaw string
	var path string
	var line int
	var daily bool
	var date string
	var status string
	var toggle bool
	var done bool
	var todo bool

	cmd := &cobra.Command{
		Use:   "task",
		Short: "Show or update a task",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if done && todo {
				return errs.New(errs.ExitValidation, "--done and --todo are mutually exclusive")
			}
			if strings.TrimSpace(status) != "" && (toggle || done || todo) {
				return errs.New(errs.ExitValidation, "--status cannot be combined with --toggle/--done/--todo")
			}

			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}

			ref, err := resolveTaskRef(rt, refRaw, path, line, daily, date)
			if err != nil {
				return err
			}
			updating := toggle || done || todo || strings.TrimSpace(status) != ""
			if updating {
				task, err := rt.Backend.UpdateTask(rt.Context, ref, tasks.UpdateInput{
					Status: status,
					Toggle: toggle,
					Done:   done,
					Todo:   todo,
				})
				if err != nil {
					return err
				}
				if rt.Printer.JSON {
					return rt.Printer.PrintJSON(task)
				}
				rt.Printer.Println(fmt.Sprintf("updated: %s:%d [%s] %s", task.Path, task.Line, task.Status, task.Text))
				return nil
			}

			task, err := rt.Backend.GetTask(rt.Context, ref)
			if err != nil {
				return err
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(task)
			}
			rt.Printer.Println(fmt.Sprintf("%s:%d [%s] %s", task.Path, task.Line, task.Status, task.Text))
			return nil
		},
	}

	cmd.Flags().StringVar(&refRaw, "ref", "", "Task ref path:line")
	cmd.Flags().StringVar(&path, "path", "", "Task file path")
	cmd.Flags().IntVar(&line, "line", 0, "Task line number")
	cmd.Flags().BoolVar(&daily, "daily", false, "Use daily note path")
	cmd.Flags().StringVar(&date, "date", "", "Daily note date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&status, "status", "", "Set custom status character")
	cmd.Flags().BoolVar(&toggle, "toggle", false, "Toggle done/todo")
	cmd.Flags().BoolVar(&done, "done", false, "Mark done ([x])")
	cmd.Flags().BoolVar(&todo, "todo", false, "Mark todo ([ ])")
	return cmd
}

func resolveTaskRef(rt *app.Runtime, refRaw, path string, line int, daily bool, date string) (tasks.Ref, error) {
	if strings.TrimSpace(refRaw) != "" {
		return tasks.ParseRef(refRaw)
	}
	if line <= 0 {
		return tasks.Ref{}, errs.New(errs.ExitValidation, "line is required when --ref is not provided")
	}

	resolvedPath := path
	if daily {
		at, err := parseDateFlag(date)
		if err != nil {
			return tasks.Ref{}, err
		}
		p, err := rt.Backend.DailyPath(rt.Context, at)
		if err != nil {
			return tasks.Ref{}, err
		}
		resolvedPath = p
	}
	if strings.TrimSpace(resolvedPath) == "" {
		return tasks.Ref{}, errs.New(errs.ExitValidation, "path is required when --ref is not provided")
	}
	return tasks.Ref{Path: resolvedPath, Line: line}, nil
}
