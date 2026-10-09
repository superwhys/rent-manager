package app

import (
	"context"
	"errors"
	"testing"

	"github.com/superwhys/rent-manager/internal/xerr"
)

func TestAccountIDFromContextPreservesAuthkitID(t *testing.T) {
	accountID := "b6ba94d9600f9cf547e7b6c20f89da0b3b176463077b3af67a91fe494c3b0e59"
	ctx := SetAccountIDToContext(context.Background(), accountID)
	got, err := AccountIDFromContext(ctx)
	if err != nil || got != accountID {
		t.Fatalf("AccountIDFromContext() = %q, %v, want %q, nil", got, err, accountID)
	}
}

func TestAccountIDFromContextRejectsMissingOrLegacyIdentity(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
	}{
		{name: "missing", ctx: context.Background()},
		{name: "empty", ctx: SetAccountIDToContext(context.Background(), "")},
		{name: "blank", ctx: SetAccountIDToContext(context.Background(), " \t\n")},
		{name: "legacy numeric ID", ctx: context.WithValue(context.Background(), accountIDContextKey{}, 10)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AccountIDFromContext(tt.ctx)
			if got != "" || !errors.Is(err, xerr.ErrUnauthorized) {
				t.Fatalf("AccountIDFromContext() = %q, %v, want empty ID and ErrUnauthorized", got, err)
			}
		})
	}
}
