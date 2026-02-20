package app

import (
	"context"
	"strings"
	"time"

	"github.com/nightisyang/obsidian-cli/internal/backend"
	"github.com/nightisyang/obsidian-cli/internal/errs"
	"github.com/nightisyang/obsidian-cli/internal/output"
	"github.com/nightisyang/obsidian-cli/internal/vault"
)

type Options struct {
	Vault   string
	Config  string
	Mode    string
	JSON    bool
	Quiet   bool
	Timeout time.Duration
}

func Build(ctx context.Context, opts Options) (*Runtime, error) {
	resolved, err := vault.ResolveConfig(opts.Vault, opts.Config)
	if err != nil {
		return nil, err
	}

	requestedMode := strings.TrimSpace(opts.Mode)
	if requestedMode == "" {
		requestedMode = resolved.Config.ModeDefault
	}
	if requestedMode == "" {
		requestedMode = "auto"
	}
	switch requestedMode {
	case "auto", "native", "api":
	default:
		return nil, errs.New(errs.ExitValidation, "mode must be one of auto, native, api")
	}

	effectiveMode := requestedMode
	if requestedMode == "auto" {
		effectiveMode = "native"
	}

	printer := output.NewPrinter(opts.JSON, opts.Quiet)

	nativeBackend := backend.NewNativeBackend(resolved.VaultRoot, resolved.Config, effectiveMode)

	return &Runtime{
		Context:       ctx,
		Config:        resolved.Config,
		VaultRoot:     resolved.VaultRoot,
		ConfigPath:    resolved.ConfigPath,
		ConfigSource:  resolved.Source,
		RequestedMode: requestedMode,
		EffectiveMode: effectiveMode,
		Printer:       printer,
		Backend:       nativeBackend,
	}, nil
}
