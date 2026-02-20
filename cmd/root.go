package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nightisyang/obsidian-cli/internal/app"
	"github.com/nightisyang/obsidian-cli/internal/errs"
	"github.com/nightisyang/obsidian-cli/internal/output"
	"github.com/spf13/cobra"
)

type runtimeKey struct{}

var rootOpts struct {
	vault   string
	config  string
	mode    string
	json    bool
	quiet   bool
	timeout time.Duration
}

func Execute() int {
	root := newRootCmd()
	err := root.Execute()
	if err == nil {
		return errs.ExitOK
	}

	code := errs.ExitCode(err)
	message := err.Error()
	if code == errs.ExitGeneric {
		if strings.Contains(message, "accepts") || strings.Contains(message, "unknown flag") || strings.Contains(message, "required flag") {
			code = errs.ExitValidation
		}
	}
	var appErr *errs.AppError
	if errors.As(err, &appErr) {
		message = appErr.Message
	}
	if rootOpts.json {
		_ = output.WriteJSON(os.Stderr, output.Failure(code, message))
	} else {
		fmt.Fprintln(os.Stderr, message)
	}
	return code
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "obsidian-cli",
		Short:         "Standalone CLI for Obsidian vaults",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			runtime, err := app.Build(cmd.Context(), app.Options{
				Vault:   rootOpts.vault,
				Config:  rootOpts.config,
				Mode:    rootOpts.mode,
				JSON:    rootOpts.json,
				Quiet:   rootOpts.quiet,
				Timeout: rootOpts.timeout,
			})
			if err != nil {
				return err
			}
			ctx := context.WithValue(cmd.Context(), runtimeKey{}, runtime)
			cmd.SetContext(ctx)
			return nil
		},
	}

	root.PersistentFlags().StringVar(&rootOpts.vault, "vault", "", "Vault root path")
	root.PersistentFlags().StringVar(&rootOpts.config, "config", "", "Config file path")
	root.PersistentFlags().StringVar(&rootOpts.mode, "mode", "", "Mode: native|api|auto")
	root.PersistentFlags().BoolVar(&rootOpts.json, "json", false, "JSON output")
	root.PersistentFlags().BoolVar(&rootOpts.quiet, "quiet", false, "Quiet human output")
	root.PersistentFlags().DurationVar(&rootOpts.timeout, "timeout", 5*time.Second, "Command timeout")

	root.AddCommand(newVaultCmd())
	root.AddCommand(newNoteCmd())
	root.AddCommand(newPropCmd())
	root.AddCommand(newTagCmd())
	root.AddCommand(newSearchCmd())
	root.AddCommand(newLinksCmd())

	return root
}

func getRuntime(cmd *cobra.Command) (*app.Runtime, error) {
	value := cmd.Context().Value(runtimeKey{})
	runtime, ok := value.(*app.Runtime)
	if !ok || runtime == nil {
		return nil, errs.New(errs.ExitGeneric, "runtime context unavailable")
	}
	return runtime, nil
}
