package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nightisyang/obsidian-cli/internal/errs"
	"github.com/spf13/cobra"
)

type listEntry struct {
	Path string `json:"path"`
	Name string `json:"name"`
	Type string `json:"type"`
}

func newListCmd() *cobra.Command {
	var includeHidden bool
	var limit int
	cmd := &cobra.Command{
		Use:   "list [path]",
		Short: "List files and folders in vault",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}

			relative := ""
			if len(args) == 1 {
				relative = args[0]
			}
			abs, rel, err := resolveVaultRelativePath(rt.VaultRoot, relative)
			if err != nil {
				return err
			}

			entries, err := os.ReadDir(abs)
			if err != nil {
				if os.IsNotExist(err) {
					return errs.New(errs.ExitNotFound, "path not found")
				}
				return err
			}

			items := make([]listEntry, 0, len(entries))
			for _, entry := range entries {
				name := entry.Name()
				if !includeHidden && strings.HasPrefix(name, ".") {
					continue
				}
				itemRel := filepath.ToSlash(filepath.Join(rel, name))
				if rel == "" {
					itemRel = filepath.ToSlash(name)
				}
				entryType := "file"
				if entry.IsDir() {
					entryType = "dir"
				}
				items = append(items, listEntry{
					Path: itemRel,
					Name: name,
					Type: entryType,
				})
			}

			sort.Slice(items, func(i, j int) bool {
				if items[i].Type != items[j].Type {
					return items[i].Type < items[j].Type
				}
				return items[i].Path < items[j].Path
			})
			if limit > 0 && len(items) > limit {
				items = items[:limit]
			}

			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(items)
			}
			for _, item := range items {
				marker := "-"
				if item.Type == "dir" {
					marker = "d"
				}
				rt.Printer.Println(fmt.Sprintf("%s\t%s", marker, item.Path))
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&includeHidden, "hidden", false, "Include hidden files and directories")
	cmd.Flags().IntVar(&limit, "limit", 0, "Limit result count")
	return cmd
}

func resolveVaultRelativePath(vaultRoot, rel string) (string, string, error) {
	cleanRoot := filepath.Clean(vaultRoot)
	targetRel := strings.TrimSpace(strings.ReplaceAll(rel, "\\", "/"))
	targetRel = strings.TrimPrefix(targetRel, "/")
	targetRel = filepath.ToSlash(filepath.Clean(targetRel))
	if targetRel == "." {
		targetRel = ""
	}

	abs := cleanRoot
	if targetRel != "" {
		abs = filepath.Join(cleanRoot, targetRel)
	}
	abs = filepath.Clean(abs)
	if abs != cleanRoot && !strings.HasPrefix(abs, cleanRoot+string(filepath.Separator)) {
		return "", "", errs.New(errs.ExitValidation, "path escapes vault root")
	}
	return abs, filepath.ToSlash(targetRel), nil
}
