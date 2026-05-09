package building

import "context"

type Repository interface {
	Create(ctx context.Context, building *Building) (*Building, error)
	Get(ctx context.Context, accountID int, id int) (*Building, error)
	List(ctx context.Context, query ListQuery) ([]*Building, error)
	Update(ctx context.Context, building *Building) (*Building, error)
	Delete(ctx context.Context, accountID int, id int) error
}
