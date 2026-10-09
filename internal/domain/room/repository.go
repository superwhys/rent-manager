package room

import "context"

type Repository interface {
	Create(ctx context.Context, accountID string, room *Room) (*Room, error)
	Get(ctx context.Context, accountID string, id int) (*Room, error)
	GetByID(ctx context.Context, id int) (*Room, error)
	List(ctx context.Context, accountID string, query ListQuery) ([]*Room, error)
	Count(ctx context.Context, accountID string, query ListQuery) (int, error)
	Update(ctx context.Context, accountID string, room *Room) (*Room, error)
	Delete(ctx context.Context, accountID string, id int) error
}
