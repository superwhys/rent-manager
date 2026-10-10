package meterreading

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	appcore "github.com/superwhys/rent-manager/internal/app"
	domainmeteritem "github.com/superwhys/rent-manager/internal/domain/meteritem"
	domainmeterreading "github.com/superwhys/rent-manager/internal/domain/meterreading"
	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
	domainroom "github.com/superwhys/rent-manager/internal/domain/room"
	domainunit "github.com/superwhys/rent-manager/internal/domain/unit"
)

// Service 提供房间抄表用例。
type Service struct {
	repo        domainmeterreading.Repository
	roomRepo    domainroom.Repository
	unitRepo    domainunit.Repository
	rbacService domainrbac.IRbacService
}

// NewService 创建房间抄表服务。
func NewService(
	repo domainmeterreading.Repository,
	roomRepo domainroom.Repository,
	unitRepo domainunit.Repository,
	rbacRepo domainrbac.Repository,
) *Service {
	return &Service{
		repo:        repo,
		roomRepo:    roomRepo,
		unitRepo:    unitRepo,
		rbacService: domainrbac.NewRbacService(rbacRepo),
	}
}

// GetRoomMonthReadings 获取房间某个月的抄表状态。
func (s *Service) GetRoomMonthReadings(
	ctx context.Context,
	query *RoomMeterReadingsQuery,
) (*RoomMonthReadings, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	periodMonth, err := validatePeriodMonth(query.PeriodMonth)
	if err != nil {
		return nil, err
	}
	room, unit, err := s.requireRoomPermission(
		ctx,
		accountID,
		query.RoomID,
		domainrbac.PermissionMeterReadingView,
	)
	if err != nil {
		return nil, err
	}
	return s.buildRoomMonthReadings(ctx, unit.AccountID, room, periodMonth)
}

// SaveRoomMonthReadings 保存房间某个月的一组抄表记录。
func (s *Service) SaveRoomMonthReadings(
	ctx context.Context,
	cmd *SaveRoomMonthCommand,
) (*RoomMonthReadings, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	periodMonth, err := validatePeriodMonth(cmd.PeriodMonth)
	if err != nil {
		return nil, err
	}
	if len(cmd.Readings) == 0 {
		return nil, domainmeterreading.ErrReadingsRequired
	}

	room, unit, err := s.requireRoomPermission(
		ctx,
		accountID,
		cmd.RoomID,
		domainrbac.PermissionMeterReadingManage,
	)
	if err != nil {
		return nil, err
	}
	itemByID := roomMeterItemMap(room)
	now := time.Now()
	seen := make(map[int]struct{}, len(cmd.Readings))
	readings := make([]*domainmeterreading.MeterReading, 0, len(cmd.Readings))

	for _, input := range cmd.Readings {
		if _, ok := seen[input.MeterItemID]; ok {
			return nil, domainmeterreading.ErrDuplicateMeterItem
		}
		seen[input.MeterItemID] = struct{}{}

		item := itemByID[input.MeterItemID]
		if item == nil {
			return nil, domainmeterreading.ErrRoomMeterItemNotAssigned
		}

		previousReading, err := s.resolvePreviousReading(
			ctx,
			unit.AccountID,
			room.ID,
			item,
			periodMonth,
			input.PreviousReading,
		)
		if err != nil {
			return nil, err
		}

		readingTime := now
		if input.ReadingTime != nil && !input.ReadingTime.IsZero() {
			readingTime = *input.ReadingTime
		}
		reading, err := domainmeterreading.NewMeterReading(domainmeterreading.CreateParams{
			RoomID:          room.ID,
			MeterItemID:     input.MeterItemID,
			PeriodMonth:     periodMonth,
			PreviousReading: previousReading,
			CurrentReading:  input.CurrentReading,
			ReadingTime:     readingTime,
			ReaderID:        &accountID,
			Source:          domainmeterreading.SourceManual,
			Remark:          input.Remark,
		})
		if err != nil {
			return nil, err
		}
		if err := reading.CalculateUsage(item.CalcType); err != nil {
			return nil, err
		}
		readings = append(readings, reading)
	}

	if _, err := s.repo.SaveRoomMonth(ctx, unit.AccountID, readings); err != nil {
		return nil, err
	}
	return s.buildRoomMonthReadings(ctx, unit.AccountID, room, periodMonth)
}

// GetUnitMonthReadings 获取单元某个月的抄表汇总。
func (s *Service) GetUnitMonthReadings(
	ctx context.Context,
	query *UnitMeterReadingsQuery,
) (*UnitMonthReadings, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	periodMonth, err := validatePeriodMonth(query.PeriodMonth)
	if err != nil {
		return nil, err
	}
	unit, err := s.requireUnitPermission(
		ctx,
		accountID,
		query.UnitID,
		domainrbac.PermissionMeterReadingView,
	)
	if err != nil {
		return nil, err
	}
	rooms, err := s.roomRepo.List(ctx, unit.AccountID, domainroom.ListQuery{UnitID: unit.ID})
	if err != nil {
		return nil, err
	}

	unitMonth := &UnitMonthReadings{
		UnitID:      unit.ID,
		PeriodMonth: periodMonth,
		Rooms:       make([]*RoomMonthReadings, 0, len(rooms)),
	}
	summaryByMeterItemID := make(map[int]*MeterItemSummary)
	for _, room := range rooms {
		roomMonth, err := s.buildRoomMonthReadings(ctx, unit.AccountID, room, periodMonth)
		if err != nil {
			return nil, err
		}
		unitMonth.Rooms = append(unitMonth.Rooms, roomMonth)
		unitMonth.TotalFee += roomMonth.TotalFee
		unitMonth.MissingCount += roomMonth.MissingCount
		collectRoomMonthSummary(summaryByMeterItemID, roomMonth)
	}

	unitMonth.Summaries = meterItemSummaries(summaryByMeterItemID)
	return unitMonth, nil
}

// buildRoomMonthReadings 构建房间某个月所有当前绑定抄表项的录入状态。
func (s *Service) buildRoomMonthReadings(
	ctx context.Context,
	accountID string,
	room *domainroom.Room,
	periodMonth string,
) (*RoomMonthReadings, error) {
	readings, err := s.repo.List(ctx, accountID, domainmeterreading.ListQuery{
		RoomID:      room.ID,
		PeriodMonth: periodMonth,
	})
	if err != nil {
		return nil, err
	}
	readingByMeterItemID := make(map[int]*domainmeterreading.MeterReading, len(readings))
	for _, reading := range readings {
		if reading == nil {
			continue
		}
		readingByMeterItemID[reading.MeterItemID] = reading
	}

	roomMonth := &RoomMonthReadings{
		Room:        room,
		PeriodMonth: periodMonth,
		Readings:    make([]*MeterReadingLine, 0, len(room.MeterItems)),
	}
	for _, item := range room.MeterItems {
		if item == nil {
			continue
		}
		line, err := s.buildMeterReadingLine(
			ctx,
			accountID,
			room.ID,
			item,
			periodMonth,
			readingByMeterItemID[item.ID],
		)
		if err != nil {
			return nil, err
		}
		roomMonth.Readings = append(roomMonth.Readings, line)
		roomMonth.TotalFee += line.Fee
		if !line.IsRecorded {
			roomMonth.MissingCount++
		}
	}
	return roomMonth, nil
}

func (s *Service) buildMeterReadingLine(
	ctx context.Context,
	accountID string,
	roomID int,
	item *domainmeteritem.MeterItem,
	periodMonth string,
	reading *domainmeterreading.MeterReading,
) (*MeterReadingLine, error) {
	line := &MeterReadingLine{
		RoomID:      roomID,
		PeriodMonth: periodMonth,
		MeterItem:   item,
		Reading:     reading,
		IsRecorded:  reading != nil,
	}
	if reading != nil {
		line.PreviousReading = reading.PreviousReading
		fee, err := reading.CalculateFee(item.UnitPrice)
		if err != nil {
			return nil, err
		}
		line.Fee = fee
		return line, nil
	}
	if item.CalcType != domainmeteritem.CalcTypeDifference {
		return line, nil
	}
	latest, err := s.repo.GetLatestBeforePeriod(ctx, accountID, roomID, item.ID, periodMonth)
	if err != nil {
		if errors.Is(err, domainmeterreading.ErrNotFound) {
			return line, nil
		}
		return nil, err
	}
	previousReading := latest.CurrentReading
	line.PreviousReading = &previousReading
	return line, nil
}

func (s *Service) resolvePreviousReading(
	ctx context.Context,
	accountID string,
	roomID int,
	item *domainmeteritem.MeterItem,
	periodMonth string,
	inputPreviousReading *float64,
) (*float64, error) {
	if item.CalcType != domainmeteritem.CalcTypeDifference {
		return nil, nil
	}
	latest, err := s.repo.GetLatestBeforePeriod(ctx, accountID, roomID, item.ID, periodMonth)
	if err == nil {
		previousReading := latest.CurrentReading
		return &previousReading, nil
	}
	if !errors.Is(err, domainmeterreading.ErrNotFound) {
		return nil, err
	}
	if inputPreviousReading == nil {
		return nil, domainmeterreading.ErrPreviousReadingRequired
	}
	return inputPreviousReading, nil
}

func (s *Service) requireRoomPermission(
	ctx context.Context,
	accountID string,
	roomID int,
	permissionCode string,
) (*domainroom.Room, *domainunit.Unit, error) {
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return nil, nil, err
	}
	unit, err := s.requireUnitPermission(ctx, accountID, room.UnitID, permissionCode)
	if err != nil {
		return nil, nil, err
	}
	return room, unit, nil
}

func (s *Service) requireUnitPermission(
	ctx context.Context,
	accountID string,
	unitID int,
	permissionCode string,
) (*domainunit.Unit, error) {
	unit, err := s.unitRepo.GetByID(ctx, unitID)
	if err != nil {
		return nil, err
	}
	if unit.AccountID == accountID {
		return unit, nil
	}
	if err := s.rbacService.RequireUnitPermission(
		ctx,
		accountID,
		unitID,
		permissionCode,
	); err != nil {
		return nil, err
	}
	return unit, nil
}

func validatePeriodMonth(periodMonth string) (string, error) {
	periodMonth = strings.TrimSpace(periodMonth)
	if periodMonth == "" {
		return "", domainmeterreading.ErrPeriodMonthRequired
	}
	if _, err := time.Parse("2006-01", periodMonth); err != nil {
		return "", domainmeterreading.ErrInvalidPeriodMonth
	}
	return periodMonth, nil
}

func roomMeterItemMap(room *domainroom.Room) map[int]*domainmeteritem.MeterItem {
	items := make(map[int]*domainmeteritem.MeterItem)
	if room == nil {
		return items
	}
	for _, item := range room.MeterItems {
		if item == nil {
			continue
		}
		items[item.ID] = item
	}
	return items
}

func collectRoomMonthSummary(
	summaryByMeterItemID map[int]*MeterItemSummary,
	roomMonth *RoomMonthReadings,
) {
	for _, line := range roomMonth.Readings {
		if line == nil || line.MeterItem == nil {
			continue
		}
		summary := summaryByMeterItemID[line.MeterItem.ID]
		if summary == nil {
			summary = &MeterItemSummary{MeterItem: line.MeterItem}
			summaryByMeterItemID[line.MeterItem.ID] = summary
		}
		if !line.IsRecorded {
			summary.Missing++
			continue
		}
		summary.Recorded++
		summary.Fee += line.Fee
		if line.Reading != nil && line.Reading.UsageAmount != nil {
			summary.UsageAmount += *line.Reading.UsageAmount
		}
	}
}

func meterItemSummaries(summaryByMeterItemID map[int]*MeterItemSummary) []*MeterItemSummary {
	summaries := make([]*MeterItemSummary, 0, len(summaryByMeterItemID))
	for _, summary := range summaryByMeterItemID {
		summaries = append(summaries, summary)
	}
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].MeterItem.ID < summaries[j].MeterItem.ID
	})
	return summaries
}
