package room

import (
	"strings"
	"time"
)

type Room struct {
	ID         int
	RoomNo     string
	Floor      int
	Price      int
	Size       int
	BuildingID int
	RoomTypeID int
	IsActive   bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewRoom(params CreateParams) (*Room, error) {
	room := &Room{
		RoomNo:     strings.TrimSpace(params.RoomNo),
		Floor:      params.Floor,
		Price:      params.Price,
		Size:       params.Size,
		BuildingID: params.BuildingID,
		RoomTypeID: params.RoomTypeID,
		IsActive:   true,
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
	r.BuildingID = params.BuildingID
	r.RoomTypeID = params.RoomTypeID
	r.IsActive = params.IsActive
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
	if r.BuildingID <= 0 {
		return ErrBuildingRequired
	}
	if r.RoomTypeID <= 0 {
		return ErrRoomTypeRequired
	}
	return nil
}
