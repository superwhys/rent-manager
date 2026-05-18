package services

import (
	"context"

	domainroomtype "github.com/superwhys/rent-manager-server/internal/domain/roomtype"
	"github.com/superwhys/rent-manager-server/internal/models/requests"
)

type RoomTypeService struct {
	repo domainroomtype.Repository
}

func NewRoomTypeService(repo domainroomtype.Repository) *RoomTypeService {
	return &RoomTypeService{repo: repo}
}

func (s *RoomTypeService) CreateRoomType(ctx context.Context, cmd *requests.CreateRoomTypeCommand) (*domainroomtype.RoomType, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	roomType, err := domainroomtype.NewRoomType(domainroomtype.CreateParams{
		AccountID: accountID,
		Name:      cmd.Name,
	})
	if err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, roomType)
}

func (s *RoomTypeService) GetRoomType(ctx context.Context, id int) (*domainroomtype.RoomType, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.repo.Get(ctx, accountID, id)
}

func (s *RoomTypeService) ListRoomTypes(ctx context.Context, query *requests.ListRoomTypesQuery) ([]*domainroomtype.RoomType, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.repo.List(ctx, domainroomtype.ListQuery{
		AccountID: accountID,
		Name:      query.Name,
		Limit:     query.Limit,
		Offset:    query.Offset,
	})
}

func (s *RoomTypeService) UpdateRoomType(ctx context.Context, cmd *requests.UpdateRoomTypeCommand) (*domainroomtype.RoomType, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	roomType := &domainroomtype.RoomType{}
	if err := roomType.Update(domainroomtype.UpdateParams{
		ID:        cmd.RoomTypeID,
		AccountID: accountID,
		Name:      cmd.Name,
	}); err != nil {
		return nil, err
	}

	return s.repo.Update(ctx, roomType)
}

func (s *RoomTypeService) DeleteRoomType(ctx context.Context, id int) error {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, accountID, id)
}
