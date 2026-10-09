package room

import (
	"context"

	appcore "github.com/superwhys/rent-manager/internal/app"
	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
	domainroom "github.com/superwhys/rent-manager/internal/domain/room"
	domainunit "github.com/superwhys/rent-manager/internal/domain/unit"
)

// Service 提供房间管理用例。
type Service struct {
	repo        domainroom.Repository
	unitRepo    domainunit.Repository
	rbacService domainrbac.IRbacService
}

// NewService 创建房间服务。
func NewService(repo domainroom.Repository, unitRepo domainunit.Repository, rbacRepo domainrbac.Repository) *Service {
	return &Service{
		repo:        repo,
		unitRepo:    unitRepo,
		rbacService: domainrbac.NewRbacService(rbacRepo),
	}
}

// CreateRoomInUnit 在当前账号可管理的单元下创建房间。
func (s *Service) CreateRoomInUnit(ctx context.Context, cmd *CreateRoomInUnitCommand) (*domainroom.Room, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	unit, err := s.requireUnitPermission(ctx, accountID, cmd.UnitID, domainrbac.PermissionRoomManage)
	if err != nil {
		return nil, err
	}

	room, err := domainroom.NewRoom(domainroom.CreateParams{
		RoomNo:       cmd.RoomNo,
		Floor:        cmd.Floor,
		Price:        cmd.Price,
		Size:         cmd.Size,
		UnitID:       cmd.UnitID,
		RoomTypeID:   cmd.RoomTypeID,
		MeterItemIDs: mergeMeterItemIDs(unit.DefaultMeterItemIDs, cmd.MeterItemIDs),
	})
	if err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, unit.AccountID, room)
}

// GetRoom 获取当前账号可访问的房间详情。
func (s *Service) GetRoom(ctx context.Context, roomID int) (*domainroom.Room, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	room, err := s.repo.GetByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	if _, err := s.requireUnitPermission(ctx, accountID, room.UnitID, domainrbac.PermissionRoomView); err != nil {
		return nil, err
	}

	return room, nil
}

// ListRoomsInUnit 获取当前账号可访问的单元房间列表。
func (s *Service) ListRoomsInUnit(ctx context.Context, query *UnitRoomsQuery) ([]*domainroom.Room, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	unit, err := s.requireUnitPermission(ctx, accountID, query.UnitID, domainrbac.PermissionRoomView)
	if err != nil {
		return nil, err
	}

	roomTypeID := 0
	if query.RoomTypeID != nil {
		roomTypeID = *query.RoomTypeID
	}

	return s.repo.List(ctx, unit.AccountID, domainroom.ListQuery{
		UnitID:     query.UnitID,
		RoomTypeID: roomTypeID,
		IsActive:   query.IsActive,
		Limit:      query.Limit,
		Offset:     query.Offset,
	})
}

// UpdateRoom 更新当前账号可管理的房间。
func (s *Service) UpdateRoom(ctx context.Context, cmd *UpdateRoomCommand) (*domainroom.Room, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	existing, err := s.repo.GetByID(ctx, cmd.RoomID)
	if err != nil {
		return nil, err
	}

	unit, err := s.requireUnitPermission(ctx, accountID, existing.UnitID, domainrbac.PermissionRoomManage)
	if err != nil {
		return nil, err
	}

	next := *existing
	if cmd.RoomNo != "" {
		next.RoomNo = cmd.RoomNo
	}
	if cmd.Floor != nil {
		next.Floor = *cmd.Floor
	}
	if cmd.Price != nil {
		next.Price = *cmd.Price
	}
	if cmd.Size != nil {
		next.Size = *cmd.Size
	}
	if cmd.UnitID != nil {
		if *cmd.UnitID != existing.UnitID {
			targetUnit, err := s.requireUnitPermission(ctx, accountID, *cmd.UnitID, domainrbac.PermissionRoomManage)
			if err != nil {
				return nil, err
			}
			if targetUnit.AccountID != unit.AccountID {
				return nil, domainroom.ErrUnitRequired
			}
		}
		next.UnitID = *cmd.UnitID
	}
	if cmd.RoomTypeID != nil {
		next.RoomTypeID = *cmd.RoomTypeID
	}
	if cmd.IsActive != nil {
		next.IsActive = *cmd.IsActive
	}
	if cmd.MeterItemIDs != nil {
		next.MeterItemIDs = cmd.MeterItemIDs
	}
	if err := next.Update(domainroom.UpdateParams{
		ID:           cmd.RoomID,
		RoomNo:       next.RoomNo,
		Floor:        next.Floor,
		Price:        next.Price,
		Size:         next.Size,
		UnitID:       next.UnitID,
		RoomTypeID:   next.RoomTypeID,
		IsActive:     next.IsActive,
		MeterItemIDs: next.MeterItemIDs,
	}); err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, unit.AccountID, &next)
}

// DeleteRoom 删除当前账号可管理的房间。
func (s *Service) DeleteRoom(ctx context.Context, roomID int) error {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return err
	}

	room, err := s.repo.GetByID(ctx, roomID)
	if err != nil {
		return err
	}

	unit, err := s.requireUnitPermission(ctx, accountID, room.UnitID, domainrbac.PermissionRoomManage)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, unit.AccountID, roomID)
}

// requireUnitPermission 校验当前账号对单元的权限。
func (s *Service) requireUnitPermission(ctx context.Context, accountID string, unitID int, permissionCode string) (*domainunit.Unit, error) {
	unit, err := s.unitRepo.GetByID(ctx, unitID)
	if err != nil {
		return nil, err
	}

	if unit.AccountID == accountID {
		return unit, nil
	}

	if err := s.rbacService.RequireUnitPermission(ctx, accountID, unitID, permissionCode); err != nil {
		return nil, err
	}

	return unit, nil
}

func mergeMeterItemIDs(defaultIDs []int, extraIDs []int) []int {
	ids := make([]int, 0, len(defaultIDs)+len(extraIDs))
	seen := make(map[int]struct{}, len(defaultIDs)+len(extraIDs))
	for _, id := range defaultIDs {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	for _, id := range extraIDs {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
}
