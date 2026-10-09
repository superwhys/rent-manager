package mysql

import (
	"context"
	"strings"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"github.com/superwhys/rent-manager/internal/infra/persistence/mysql/query"
)

func accountRooms(ctx context.Context, q *query.Query, accountID int) query.IRoomDo {
	u := q.Unit
	r := q.Room
	return r.WithContext(ctx).Where(
		r.DeletedAt.IsNull(),
		gen.Columns{r.UnitID}.In(u.WithContext(ctx).Select(u.ID).Where(u.AccountID.Eq(accountID), u.DeletedAt.IsNull())),
	)
}

// contains preserves literal %, _ and backslash searches; Gen Like has no ESCAPE option.
func contains(column field.String, value string) gen.Condition {
	pattern := "%" + strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_").Replace(value) + "%"
	return field.NewUnsafeFieldRaw("? LIKE ? ESCAPE ?", column.RawExpr(), pattern, "\\")
}
