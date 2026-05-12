package room

import "context"

type Repository interface {
	Create(ctx context.Context, organizationID int, room *Room) (*Room, error)
	Get(ctx context.Context, organizationID int, id int) (*Room, error)
	List(ctx context.Context, organizationID int, query ListQuery) ([]*Room, error)
	Update(ctx context.Context, organizationID int, room *Room) (*Room, error)
	Delete(ctx context.Context, organizationID int, id int) error
}
