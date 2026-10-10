package rentbill

import "context"

type Repository interface {
	Create(ctx context.Context, accountID string, bill *RentBill) (*RentBill, error)
	Get(ctx context.Context, accountID string, id int) (*RentBill, error)
	GetByID(ctx context.Context, id int) (*RentBill, error)
	GetByRoomPeriod(
		ctx context.Context,
		accountID string,
		roomID int,
		periodMonth string,
	) (*RentBill, error)
	List(ctx context.Context, accountID string, query ListQuery) ([]*RentBill, error)
	SaveGeneratedBills(
		ctx context.Context,
		accountID string,
		bills []*RentBill,
	) (*SaveGeneratedBillsResult, error)
	Update(ctx context.Context, accountID string, bill *RentBill) (*RentBill, error)
	Delete(ctx context.Context, accountID string, id int) error
}
