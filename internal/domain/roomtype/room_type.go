package roomtype

import (
	"strings"
	"time"
)

const maxNameLength = 12

type RoomType struct {
	ID        int
	AccountID string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func NewRoomType(params CreateParams) (*RoomType, error) {
	roomType := &RoomType{
		AccountID: params.AccountID,
		Name:      strings.TrimSpace(params.Name),
	}
	if err := roomType.Validate(); err != nil {
		return nil, err
	}
	return roomType, nil
}

func (r *RoomType) Update(params UpdateParams) error {
	r.ID = params.ID
	r.AccountID = params.AccountID
	r.Name = strings.TrimSpace(params.Name)
	return r.Validate()
}

func (r *RoomType) Validate() error {
	if strings.TrimSpace(r.AccountID) == "" {
		return ErrAccountRequired
	}
	if r.Name == "" {
		return ErrNameRequired
	}
	if len([]rune(r.Name)) > maxNameLength {
		return ErrNameTooLong
	}
	return nil
}
