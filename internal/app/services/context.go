package services

import (
	"context"

	"git.superwhys.top/rent-manager/rent-manager-server/internal/jwtutils"
)

func AccountIDFromContext(ctx context.Context) (int, error) {
	claims, err := jwtutils.TokenClaimsFromContext(ctx)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}
