package roomtype

import "context"

type Repository interface {
	Create(ctx context.Context, roomType *RoomType) (*RoomType, error)
	Get(ctx context.Context, accountID int, id int) (*RoomType, error)
	List(ctx context.Context, query ListQuery) ([]*RoomType, error)
	Update(ctx context.Context, roomType *RoomType) (*RoomType, error)
	Delete(ctx context.Context, accountID int, id int) error
}
