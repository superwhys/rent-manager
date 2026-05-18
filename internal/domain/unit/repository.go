package unit

import "context"

type Repository interface {
	Create(ctx context.Context, unit *Unit) (*Unit, error)
	Get(ctx context.Context, accountID int, id int) (*Unit, error)
	GetByID(ctx context.Context, id int) (*Unit, error)
	List(ctx context.Context, query ListQuery) ([]*Unit, error)
	Update(ctx context.Context, unit *Unit) (*Unit, error)
	Delete(ctx context.Context, accountID int, id int) error
}
