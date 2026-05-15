package services

import (
	"context"
	"errors"

	domainmeteritem "github.com/superwhys/rent-manager-server/internal/domain/meteritem"
	"github.com/superwhys/rent-manager-server/internal/errcode"
	"github.com/superwhys/rent-manager-server/internal/models/requests"
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
		CalcType:  cmd.CalcType,
		Remark:    cmd.Remark,
	})
	if err != nil {
		return nil, mapMeterItemServiceError(err)
	}

	item, err = s.repo.Create(ctx, item)
	if err != nil {
		return nil, mapMeterItemServiceError(err)
	}
	return item, nil
}

func (s *MeterItemService) ListMeterItems(ctx context.Context, query *requests.ListMeterItemsQuery) ([]*domainmeteritem.MeterItem, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.List(ctx, domainmeteritem.ListQuery{
		AccountID: accountID,
		Code:      query.Code,
		IsSystem:  query.IsSystem,
		CalcType:  query.CalcType,
		Limit:     query.Limit,
		Offset:    query.Offset,
	})
	if err != nil {
		return nil, mapMeterItemServiceError(err)
	}
	return items, nil
}

func (s *MeterItemService) UpdateMeterItem(ctx context.Context, cmd *requests.UpdateMeterItemCommand) (*domainmeteritem.MeterItem, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	item, err := s.repo.Get(ctx, accountID, cmd.MeterItemID)
	if err != nil {
		return nil, mapMeterItemServiceError(err)
	}
	if err := item.Update(domainmeteritem.UpdateParams{
		ID:        cmd.MeterItemID,
		AccountID: accountID,
		Code:      cmd.Code,
		Unit:      cmd.Unit,
		CalcType:  cmd.CalcType,
		Remark:    cmd.Remark,
	}); err != nil {
		return nil, mapMeterItemServiceError(err)
	}

	item, err = s.repo.Update(ctx, item)
	if err != nil {
		return nil, mapMeterItemServiceError(err)
	}
	return item, nil
}

func mapMeterItemServiceError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, domainmeteritem.ErrNotFound):
		return errcode.ErrNotFound.WithMessage(err.Error())
	case errors.Is(err, domainmeteritem.ErrAccountRequired),
		errors.Is(err, domainmeteritem.ErrCodeRequired),
		errors.Is(err, domainmeteritem.ErrUnitRequired),
		errors.Is(err, domainmeteritem.ErrInvalidCalcType),
		errors.Is(err, domainmeteritem.ErrCodeAlreadyExists):
		return errcode.ErrBadRequest.WithMessage(err.Error())
	default:
		return err
	}
}
