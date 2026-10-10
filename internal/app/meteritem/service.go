package meteritem

import (
	"context"

	appcore "github.com/superwhys/rent-manager/internal/app"
	domainmeteritem "github.com/superwhys/rent-manager/internal/domain/meteritem"
	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
	domainunit "github.com/superwhys/rent-manager/internal/domain/unit"
)

// Service 提供抄表项管理用例。
type Service struct {
	repo        domainmeteritem.Repository
	unitRepo    domainunit.Repository
	rbacService domainrbac.IRbacService
}

// NewService 创建抄表项服务。
func NewService(
	repo domainmeteritem.Repository,
	unitRepo domainunit.Repository,
	rbacRepo domainrbac.Repository,
) *Service {
	return &Service{
		repo:        repo,
		unitRepo:    unitRepo,
		rbacService: domainrbac.NewRbacService(rbacRepo),
	}
}

// CreateMeterItem 创建当前账号下的抄表项。
func (s *Service) CreateMeterItem(
	ctx context.Context,
	cmd *CreateMeterItemCommand,
) (*domainmeteritem.MeterItem, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	item, err := domainmeteritem.NewMeterItem(domainmeteritem.CreateParams{
		AccountID: accountID,
		Code:      cmd.Code,
		Unit:      cmd.Unit,
		UnitPrice: cmd.UnitPrice,
		CalcType:  cmd.CalcType,
		Remark:    cmd.Remark,
	})
	if err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, item)
}

// GetMeterItem 获取当前账号或授权单元所有者的抄表项详情。
func (s *Service) GetMeterItem(ctx context.Context, query *GetMeterItemQuery) (*domainmeteritem.MeterItem, error) {
	accountID, err := s.readAccountID(ctx, query.UnitID)
	if err != nil {
		return nil, err
	}

	return s.repo.Get(ctx, accountID, query.MeterItemID)
}

// ListMeterItems 获取当前账号或授权单元所有者的抄表项列表。
func (s *Service) ListMeterItems(
	ctx context.Context,
	query *ListMeterItemsQuery,
) ([]*domainmeteritem.MeterItem, error) {
	accountID, err := s.readAccountID(ctx, query.UnitID)
	if err != nil {
		return nil, err
	}

	return s.repo.List(ctx, domainmeteritem.ListQuery{
		AccountID: accountID,
		Code:      query.Code,
		IsSystem:  query.IsSystem,
		CalcType:  query.CalcType,
		Limit:     query.Limit,
		Offset:    query.Offset,
	})
}

// UpdateMeterItem 更新当前账号下的抄表项。
func (s *Service) UpdateMeterItem(
	ctx context.Context,
	cmd *UpdateMeterItemCommand,
) (*domainmeteritem.MeterItem, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	item, err := domainmeteritem.NewMeterItem(domainmeteritem.CreateParams{
		AccountID: accountID,
		Code:      cmd.Code,
		Unit:      cmd.Unit,
		UnitPrice: cmd.UnitPrice,
		CalcType:  cmd.CalcType,
		Remark:    cmd.Remark,
	})
	if err != nil {
		return nil, err
	}

	item.ID = cmd.MeterItemID
	return s.repo.Update(ctx, item)
}

func (s *Service) readAccountID(ctx context.Context, unitID int) (string, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil || unitID == 0 {
		return accountID, err
	}
	if err := s.rbacService.RequireUnitPermission(ctx, accountID, unitID, domainrbac.PermissionMeterItemView); err != nil {
		return "", err
	}
	unit, err := s.unitRepo.GetByID(ctx, unitID)
	if err != nil {
		return "", err
	}
	// 收费定义在所有者的楼栋间共享，单元授权只开放目录读取和选用。
	return unit.AccountID, nil
}
