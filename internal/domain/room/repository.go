package room

import "context"

type Repository interface {
	Create(ctx context.Context, accountID int, room *Room) (*Room, error)
	Get(ctx context.Context, accountID int, id int) (*Room, error)
	GetByID(ctx context.Context, id int) (*Room, error)
	List(ctx context.Context, accountID int, query ListQuery) ([]*Room, error)
	Update(ctx context.Context, accountID int, room *Room) (*Room, error)
	Delete(ctx context.Context, accountID int, id int) error
}
