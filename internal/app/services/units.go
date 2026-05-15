package services

import (
	"context"

	domainroom "github.com/superwhys/rent-manager-server/internal/domain/room"
	domainunit "github.com/superwhys/rent-manager-server/internal/domain/unit"
	"github.com/superwhys/rent-manager-server/internal/models/requests"
)

type UnitService struct {
	repo        domainunit.Repository
	roomService *RoomService
}

func NewUnitService(repo domainunit.Repository, roomService *RoomService) *UnitService {
	return &UnitService{repo: repo, roomService: roomService}
}

func (s *UnitService) CreateUnit(ctx context.Context, cmd *requests.CreateUnitCommand) (*domainunit.Unit, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	unit, err := domainunit.NewUnit(domainunit.CreateParams{
		AccountID:           accountID,
		Name:                cmd.Name,
		Address:             cmd.Address,
		City:                cmd.City,
		Province:            cmd.Province,
		DefaultMeterItemIDs: cmd.DefaultMeterItemIDs,
	})
	if err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, unit)
}

func (s *UnitService) ListUnits(ctx context.Context, query *requests.ListUnitsQuery) ([]*domainunit.Unit, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.repo.List(ctx, domainunit.ListQuery{
		AccountID: accountID,
		Name:      query.Name,
		City:      query.City,
		Province:  query.Province,
		Limit:     query.Limit,
		Offset:    query.Offset,
	})
}

func (s *UnitService) GetUnit(ctx context.Context, id int) (*domainunit.Unit, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, accountID, id)
}

func (s *UnitService) UpdateUnit(ctx context.Context, cmd *requests.UpdateUnitCommand) (*domainunit.Unit, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	unit, err := domainunit.NewUnit(domainunit.CreateParams{
		AccountID:           accountID,
		Name:                cmd.Name,
		Address:             cmd.Address,
		City:                cmd.City,
		Province:            cmd.Province,
		DefaultMeterItemIDs: cmd.DefaultMeterItemIDs,
	})
	if err != nil {
		return nil, err
	}

	if err := unit.Update(domainunit.UpdateParams{
		ID:                  cmd.UnitID,
		AccountID:           accountID,
		Name:                cmd.Name,
		Address:             cmd.Address,
		City:                cmd.City,
		Province:            cmd.Province,
		DefaultMeterItemIDs: cmd.DefaultMeterItemIDs,
	}); err != nil {
		return nil, err
	}

	return s.repo.Update(ctx, unit)
}

func (s *UnitService) DeleteUnit(ctx context.Context, id int) error {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, accountID, id)
}

func (s *UnitService) CreateRoomInUnit(ctx context.Context, cmd *requests.CreateRoomInUnitCommand) (*domainroom.Room, error) {
	if _, err := s.GetUnit(ctx, cmd.UnitID); err != nil {
		return nil, err
	}

	return s.roomService.CreateRoom(ctx, &requests.CreateRoomCommand{
		RoomNo:       cmd.RoomNo,
		Floor:        cmd.Floor,
		Price:        cmd.Price,
		Size:         cmd.Size,
		UnitID:       cmd.UnitID,
		RoomTypeID:   cmd.RoomTypeID,
		MeterItemIDs: cmd.MeterItemIDs,
	})
}
