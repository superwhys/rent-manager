package services

import (
	"context"

	domainbuilding "github.com/superwhys/rent-manager-server/internal/domain/building"
	domainroom "github.com/superwhys/rent-manager-server/internal/domain/room"
	"github.com/superwhys/rent-manager-server/internal/models/requests"
)

type BuildingService struct {
	repo        domainbuilding.Repository
	roomService *RoomService
}

func NewBuildingService(repo domainbuilding.Repository, roomService *RoomService) *BuildingService {
	return &BuildingService{repo: repo, roomService: roomService}
}

func (s *BuildingService) CreateBuilding(ctx context.Context, cmd *requests.CreateBuildingCommand) (*domainbuilding.Building, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	building, err := domainbuilding.NewBuilding(domainbuilding.CreateParams{
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

	building, err = s.repo.Create(ctx, building)
	if err != nil {
		return nil, err
	}

	return building, nil
}

func (s *BuildingService) ListBuildings(ctx context.Context, query *requests.ListBuildingsQuery) ([]*domainbuilding.Building, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.repo.List(ctx, domainbuilding.ListQuery{
		AccountID: accountID,
		Name:      query.Name,
		City:      query.City,
		Province:  query.Province,
		Limit:     query.Limit,
		Offset:    query.Offset,
	})
}

func (s *BuildingService) GetBuilding(ctx context.Context, id int) (*domainbuilding.Building, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	building, err := s.repo.Get(ctx, accountID, id)
	if err != nil {
		return nil, err
	}

	return building, nil
}

func (s *BuildingService) UpdateBuilding(ctx context.Context, cmd *requests.UpdateBuildingCommand) (*domainbuilding.Building, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	building, err := domainbuilding.NewBuilding(domainbuilding.CreateParams{
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

	err = building.Update(domainbuilding.UpdateParams{
		ID:                  cmd.BuildingID,
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

	building, err = s.repo.Update(ctx, building)
	if err != nil {
		return nil, err
	}

	return building, nil
}

func (s *BuildingService) DeleteBuilding(ctx context.Context, id int) error {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, accountID, id)
}

func (s *BuildingService) CreateRoomInBuilding(ctx context.Context, cmd *requests.CreateRoomInBuildingCommand) (*domainroom.Room, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if _, err := s.repo.Get(ctx, accountID, cmd.BuildingID); err != nil {
		return nil, err
	}

	room, err := s.roomService.CreateRoom(ctx, &requests.CreateRoomCommand{
		RoomNo:     cmd.RoomNo,
		Floor:      cmd.Floor,
		Price:      cmd.Price,
		Size:       cmd.Size,
		BuildingID: cmd.BuildingID,
		RoomTypeID: cmd.RoomTypeID,
	})
	if err != nil {
		return nil, err
	}

	return room, nil
}
