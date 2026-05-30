package services

import (
	"context"

	domainrbac "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/rbac"
	domainroom "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/room"
	domainunit "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/unit"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/models/requests"
)

type RoomService struct {
	repo     domainroom.Repository
	unitRepo domainunit.Repository
	rbac     *RBACService
}

func NewRoomService(repo domainroom.Repository, unitRepo domainunit.Repository, rbac *RBACService) *RoomService {
	return &RoomService{repo: repo, unitRepo: unitRepo, rbac: rbac}
}

func (s *RoomService) CreateRoomInUnit(ctx context.Context, cmd *requests.CreateRoomInUnitCommand) (*domainroom.Room, error) {
	accountID, err := AccountIDFromContext(ctx)
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
		MeterItemIDs: cmd.MeterItemIDs,
	})
	if err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, unit.AccountID, room)
}

func (s *RoomService) GetRoom(ctx context.Context, roomID int) (*domainroom.Room, error) {
	accountID, err := AccountIDFromContext(ctx)
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

func (s *RoomService) ListRoomsInUnit(ctx context.Context, query *requests.UnitRoomsQuery) ([]*domainroom.Room, error) {
	accountID, err := AccountIDFromContext(ctx)
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

func (s *RoomService) UpdateRoom(ctx context.Context, cmd *requests.UpdateRoomCommand) (*domainroom.Room, error) {
	accountID, err := AccountIDFromContext(ctx)
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

func (s *RoomService) DeleteRoom(ctx context.Context, roomID int) error {
	accountID, err := AccountIDFromContext(ctx)
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

func (s *RoomService) requireUnitPermission(ctx context.Context, accountID int, unitID int, permissionCode string) (*domainunit.Unit, error) {
	unit, err := s.unitRepo.GetByID(ctx, unitID)
	if err != nil {
		return nil, err
	}

	if unit.AccountID == accountID {
		return unit, nil
	}

	if err := s.rbac.RequireUnitPermission(ctx, accountID, unitID, permissionCode); err != nil {
		return nil, err
	}

	return unit, nil
}
