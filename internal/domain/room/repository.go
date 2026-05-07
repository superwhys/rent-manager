package room

import "context"

type Repository interface {
	Create(ctx context.Context, room *Room) (*Room, error)
	Get(ctx context.Context, id int) (*Room, error)
	List(ctx context.Context, query ListQuery) ([]*Room, error)
	Update(ctx context.Context, room *Room) (*Room, error)
	Delete(ctx context.Context, id int) error
}
