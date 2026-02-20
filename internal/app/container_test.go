package app

import (
	"context"
	"testing"

	"github.com/nightisyang/obsidian-cli/internal/errs"
)

func TestBuildRejectsInvalidMode(t *testing.T) {
	_, err := Build(context.Background(), Options{
		Vault: t.TempDir(),
		Mode:  "invalid",
	})
	if err == nil {
		t.Fatalf("expected validation error for invalid mode")
	}
	appErr, ok := err.(*errs.AppError)
	if !ok || appErr.Code != errs.ExitValidation {
		t.Fatalf("expected validation app error, got %T (%v)", err, err)
	}
}
