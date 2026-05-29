package services

import (
	"context"

	domainroom "github.com/superwhys/rent-manager-server/internal/domain/room"
	domainunit "github.com/superwhys/rent-manager-server/internal/domain/unit"
)

// HomeSummary 描述首页楼栋与房间汇总信息。
type HomeSummary struct {
	BuildingCount   int
	RoomCount       int
	RentedRoomCount int
	VacantRoomCount int
}

// HomeService 提供首页聚合数据用例。
type HomeService struct {
	unitRepo domainunit.Repository
	roomRepo domainroom.Repository
}

// NewHomeService 创建首页聚合服务。
func NewHomeService(unitRepo domainunit.Repository, roomRepo domainroom.Repository) *HomeService {
	return &HomeService{unitRepo: unitRepo, roomRepo: roomRepo}
}

// GetHomeSummary 获取当前账号可访问的首页楼栋与房间汇总。
func (s *HomeService) GetHomeSummary(ctx context.Context) (*HomeSummary, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	units, err := s.unitRepo.List(ctx, domainunit.ListQuery{
		AccountID: accountID,
		Scope:     domainunit.ListScopeAll,
	})
	if err != nil {
		return nil, err
	}

	summary := &HomeSummary{}
	active := true
	inactive := false
	for _, unit := range units {
		if unit == nil {
			continue
		}
		summary.BuildingCount++

		// 当前尚未建租约模型，出租/空置沿用房间启用状态统计。
		rentedCount, err := s.countRoomsInUnit(ctx, unit, &active)
		if err != nil {
			return nil, err
		}
		vacantCount, err := s.countRoomsInUnit(ctx, unit, &inactive)
		if err != nil {
			return nil, err
		}

		summary.RentedRoomCount += rentedCount
		summary.VacantRoomCount += vacantCount
		summary.RoomCount += rentedCount + vacantCount
	}

	return summary, nil
}

func (s *HomeService) countRoomsInUnit(ctx context.Context, unit *domainunit.Unit, isActive *bool) (int, error) {
	return s.roomRepo.Count(ctx, unit.AccountID, domainroom.ListQuery{
		UnitID:   unit.ID,
		IsActive: isActive,
	})
}
