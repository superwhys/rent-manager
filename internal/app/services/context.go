package services

import (
	"context"

	"github.com/superwhys/rent-manager-server/internal/jwtutils"
)

func AccountIDFromContext(ctx context.Context) (int, error) {
	claims, err := jwtutils.TokenClaimsFromContext(ctx)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}
