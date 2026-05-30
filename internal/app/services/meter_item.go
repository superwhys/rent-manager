package services

import (
	"context"

	domainmeteritem "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/meteritem"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/models/requests"
)

type MeterItemService struct {
	repo domainmeteritem.Repository
}

func NewMeterItemService(repo domainmeteritem.Repository) *MeterItemService {
	return &MeterItemService{repo: repo}
}

func (s *MeterItemService) CreateMeterItem(ctx context.Context, cmd *requests.CreateMeterItemCommand) (*domainmeteritem.MeterItem, error) {
	accountID, err := AccountIDFromContext(ctx)
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

func (s *MeterItemService) GetMeterItem(ctx context.Context, id int) (*domainmeteritem.MeterItem, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.repo.Get(ctx, accountID, id)
}

func (s *MeterItemService) ListMeterItems(ctx context.Context, query *requests.ListMeterItemsQuery) ([]*domainmeteritem.MeterItem, error) {
	accountID, err := AccountIDFromContext(ctx)
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

func (s *MeterItemService) UpdateMeterItem(ctx context.Context, cmd *requests.UpdateMeterItemCommand) (*domainmeteritem.MeterItem, error) {
	accountID, err := AccountIDFromContext(ctx)
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
