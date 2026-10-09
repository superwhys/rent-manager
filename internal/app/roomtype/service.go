package roomtype

import (
	"context"

	appcore "github.com/superwhys/rent-manager/internal/app"
	domainroomtype "github.com/superwhys/rent-manager/internal/domain/roomtype"
)

// Service 提供房型管理用例。
type Service struct {
	repo domainroomtype.Repository
}

// NewService 创建房型服务。
func NewService(repo domainroomtype.Repository) *Service {
	return &Service{repo: repo}
}

// CreateRoomType 创建当前账号下的房型。
func (s *Service) CreateRoomType(ctx context.Context, cmd *CreateRoomTypeCommand) (*domainroomtype.RoomType, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
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

// GetRoomType 获取当前账号下的房型详情。
func (s *Service) GetRoomType(ctx context.Context, id int) (*domainroomtype.RoomType, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.repo.Get(ctx, accountID, id)
}

// ListRoomTypes 获取当前账号下的房型列表。
func (s *Service) ListRoomTypes(ctx context.Context, query *ListRoomTypesQuery) ([]*domainroomtype.RoomType, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
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

// UpdateRoomType 更新当前账号下的房型。
func (s *Service) UpdateRoomType(ctx context.Context, cmd *UpdateRoomTypeCommand) (*domainroomtype.RoomType, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
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

// DeleteRoomType 删除当前账号下的房型。
func (s *Service) DeleteRoomType(ctx context.Context, id int) error {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, accountID, id)
}
