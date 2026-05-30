package room

import (
	"strings"
	"time"

	domainmeteritem "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/meteritem"
)

type Room struct {
	ID           int
	RoomNo       string
	Floor        int
	Price        int
	Size         int
	UnitID       int
	RoomTypeID   int
	IsActive     bool
	MeterItemIDs []int
	MeterItems   []*domainmeteritem.MeterItem
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewRoom(params CreateParams) (*Room, error) {
	room := &Room{
		RoomNo:       strings.TrimSpace(params.RoomNo),
		Floor:        params.Floor,
		Price:        params.Price,
		Size:         params.Size,
		UnitID:       params.UnitID,
		RoomTypeID:   params.RoomTypeID,
		IsActive:     true,
		MeterItemIDs: normalizeMeterItemIDs(params.MeterItemIDs),
	}
	if err := room.Validate(); err != nil {
		return nil, err
	}
	return room, nil
}

func (r *Room) Update(params UpdateParams) error {
	r.RoomNo = strings.TrimSpace(params.RoomNo)
	r.Floor = params.Floor
	r.Price = params.Price
	r.Size = params.Size
	r.UnitID = params.UnitID
	r.RoomTypeID = params.RoomTypeID
	r.IsActive = params.IsActive
	r.MeterItemIDs = normalizeMeterItemIDs(params.MeterItemIDs)
	return r.Validate()
}

func (r *Room) Validate() error {
	if r.RoomNo == "" {
		return ErrRoomNoRequired
	}
	if r.Floor < 0 {
		return ErrInvalidFloor
	}
	if r.Price < 0 {
		return ErrInvalidPrice
	}
	if r.Size <= 0 {
		return ErrInvalidSize
	}
	if r.UnitID <= 0 {
		return ErrUnitRequired
	}
	if r.RoomTypeID <= 0 {
		return ErrRoomTypeRequired
	}
	for _, id := range r.MeterItemIDs {
		if id <= 0 {
			return ErrInvalidMeterItemID
		}
	}
	return nil
}

func normalizeMeterItemIDs(ids []int) []int {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[int]struct{}, len(ids))
	normalized := make([]int, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		normalized = append(normalized, id)
	}
	return normalized
}
