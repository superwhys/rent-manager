package mysql

import (
	"context"

	"gorm.io/gorm"

	"github.com/superwhys/rent-manager/internal/domain/account"
	"github.com/superwhys/rent-manager/internal/domain/accountsetting"
	"github.com/superwhys/rent-manager/internal/domain/meteritem"
	"github.com/superwhys/rent-manager/internal/domain/meterreading"
	"github.com/superwhys/rent-manager/internal/domain/rbac"
	"github.com/superwhys/rent-manager/internal/domain/rentbill"
	"github.com/superwhys/rent-manager/internal/domain/room"
	"github.com/superwhys/rent-manager/internal/domain/roomtype"
	"github.com/superwhys/rent-manager/internal/domain/unit"
)

// Repositories groups repositories bound to the same database connection.
type Repositories struct {
	Account        account.Repository
	AccountSetting accountsetting.Repository
	RBAC           rbac.Repository
	Unit           unit.Repository
	Room           room.Repository
	RoomType       roomtype.Repository
	MeterItem      meteritem.Repository
	MeterReading   meterreading.Repository
	RentBill       rentbill.Repository
}

type RepositoryFactory struct {
	db *gorm.DB
}

func NewRepositoryFactory(db *gorm.DB) *RepositoryFactory {
	return &RepositoryFactory{db: db}
}

func (f *RepositoryFactory) Repositories() Repositories {
	return Repositories{
		Account:        NewAccountRepository(f.db),
		AccountSetting: NewAccountSettingRepository(f.db),
		RBAC:           NewRBACRepository(f.db),
		Unit:           NewUnitRepository(f.db),
		Room:           NewRoomRepository(f.db),
		RoomType:       NewRoomTypeRepository(f.db),
		MeterItem:      NewMeterItemRepository(f.db),
		MeterReading:   NewMeterReadingRepository(f.db),
		RentBill:       NewRentBillRepository(f.db),
	}
}

// WithTransaction runs all supplied repositories on the same transaction.
func (f *RepositoryFactory) WithTransaction(ctx context.Context, fn func(Repositories) error) error {
	return f.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(NewRepositoryFactory(tx).Repositories())
	})
}
