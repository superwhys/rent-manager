package roomtype

import (
	"context"

	appcore "github.com/superwhys/rent-manager/internal/app"
	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
	domainroomtype "github.com/superwhys/rent-manager/internal/domain/roomtype"
	domainunit "github.com/superwhys/rent-manager/internal/domain/unit"
)

// Service 提供房型管理用例。
type Service struct {
	repo        domainroomtype.Repository
	unitRepo    domainunit.Repository
	rbacService domainrbac.IRbacService
}

// NewService 创建房型服务。
func NewService(
	repo domainroomtype.Repository,
	unitRepo domainunit.Repository,
	rbacRepo domainrbac.Repository,
) *Service {
	return &Service{
		repo:        repo,
		unitRepo:    unitRepo,
		rbacService: domainrbac.NewRbacService(rbacRepo),
	}
}

// CreateRoomType 创建当前账号下的房型。
func (s *Service) CreateRoomType(
	ctx context.Context,
	cmd *CreateRoomTypeCommand,
) (*domainroomtype.RoomType, error) {
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

// GetRoomType 获取当前账号或授权单元所有者的房型详情。
func (s *Service) GetRoomType(ctx context.Context, query *GetRoomTypeQuery) (*domainroomtype.RoomType, error) {
	accountID, err := s.readAccountID(ctx, query.UnitID)
	if err != nil {
		return nil, err
	}

	return s.repo.Get(ctx, accountID, query.RoomTypeID)
}

// ListRoomTypes 获取当前账号或授权单元所有者的房型列表。
func (s *Service) ListRoomTypes(
	ctx context.Context,
	query *ListRoomTypesQuery,
) ([]*domainroomtype.RoomType, error) {
	accountID, err := s.readAccountID(ctx, query.UnitID)
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
func (s *Service) UpdateRoomType(
	ctx context.Context,
	cmd *UpdateRoomTypeCommand,
) (*domainroomtype.RoomType, error) {
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

func (s *Service) readAccountID(ctx context.Context, unitID int) (string, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil || unitID == 0 {
		return accountID, err
	}
	if err := s.rbacService.RequireUnitPermission(ctx, accountID, unitID, domainrbac.PermissionRoomTypeView); err != nil {
		return "", err
	}
	unit, err := s.unitRepo.GetByID(ctx, unitID)
	if err != nil {
		return "", err
	}
	// 房型是账号共享定义；授权允许选用所有者的完整目录，修改仍限定义所有者。
	return unit.AccountID, nil
}
