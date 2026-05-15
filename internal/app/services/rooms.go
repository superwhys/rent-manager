package services

import (
	"context"

	"github.com/miebyte/goutils/utils/ptrx"
	"github.com/superwhys/rent-manager-server/internal/models/requests"

	domainroom "github.com/superwhys/rent-manager-server/internal/domain/room"
)

type RoomService struct {
	repo domainroom.Repository
}

func NewRoomService(repo domainroom.Repository) *RoomService {
	return &RoomService{repo: repo}
}

func (s *RoomService) CreateRoom(ctx context.Context, cmd *requests.CreateRoomCommand) (*domainroom.Room, error) {
	accountID, err := AccountIDFromContext(ctx)
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
	return s.repo.Create(ctx, accountID, room)
}

func (s *RoomService) GetRoom(ctx context.Context, id int) (*domainroom.Room, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, accountID, id)
}

func (s *RoomService) ListRooms(ctx context.Context, query *requests.ListRoomsQuery) ([]*domainroom.Room, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.repo.List(ctx, accountID, domainroom.ListQuery{
		UnitID:     query.UnitID,
		RoomTypeID: ptrx.IntValue(query.RoomTypeID),
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
	room, err := s.repo.Get(ctx, accountID, cmd.RoomID)
	if err != nil {
		return nil, err
	}
	if err := room.Update(domainroom.UpdateParams{
		ID:           cmd.RoomID,
		RoomNo:       cmd.RoomNo,
		Floor:        cmd.Floor,
		Price:        cmd.Price,
		Size:         cmd.Size,
		UnitID:       cmd.UnitID,
		RoomTypeID:   cmd.RoomTypeID,
		IsActive:     cmd.IsActive,
		MeterItemIDs: cmd.MeterItemIDs,
	}); err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, accountID, room)
}

func (s *RoomService) DeleteRoom(ctx context.Context, id int) error {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, accountID, id)
}
