package search

import (
	"context"
	"testing"

	"github.com/nightisyang/obsidian-cli/internal/errs"
)

func TestRGEngineRejectsEscapingPath(t *testing.T) {
	engine := &RGEngine{VaultRoot: t.TempDir()}
	_, err := engine.Search(context.Background(), Query{
		Type: QueryText,
		Text: "hello",
		Path: "../outside",
	})
	if err == nil {
		t.Fatalf("expected validation error for escaping search path")
	}
	appErr, ok := err.(*errs.AppError)
	if !ok || appErr.Code != errs.ExitValidation {
		t.Fatalf("expected validation app error, got %T (%v)", err, err)
	}
}
