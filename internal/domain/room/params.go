package room

type CreateParams struct {
	RoomNo       string
	Floor        int
	Price        int
	Size         int
	UnitID       int
	RoomTypeID   int
	MeterItemIDs []int
}

type UpdateParams struct {
	ID           int
	RoomNo       string
	Floor        int
	Price        int
	Size         int
	UnitID       int
	RoomTypeID   int
	IsActive     bool
	MeterItemIDs []int
}

type ListQuery struct {
	UnitID     int
	RoomTypeID int
	IsActive   *bool
	Limit      int
	Offset     int
}
