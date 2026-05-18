package rentbill

import "context"

type Repository interface {
	Create(ctx context.Context, accountID int, bill *RentBill) (*RentBill, error)
	Get(ctx context.Context, accountID int, id int) (*RentBill, error)
	GetByRoomPeriod(ctx context.Context, accountID int, roomID int, periodMonth string) (*RentBill, error)
	List(ctx context.Context, accountID int, query ListQuery) ([]*RentBill, error)
	Update(ctx context.Context, accountID int, bill *RentBill) (*RentBill, error)
	Delete(ctx context.Context, accountID int, id int) error
}
