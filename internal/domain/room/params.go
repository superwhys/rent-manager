package room

// CreateParams 描述创建房间时需要满足的领域输入。
type CreateParams struct {
	RoomNo     string
	Floor      int
	Price      int
	Size       int
	BuildingID int
	RoomTypeID int
}

// UpdateParams 描述更新房间时允许变更的领域输入。
type UpdateParams struct {
	ID         int
	RoomNo     string
	Floor      int
	Price      int
	Size       int
	BuildingID int
	RoomTypeID int
	IsActive   bool
}

// ListQuery 描述房间列表查询条件。
type ListQuery struct {
	BuildingID int
	RoomTypeID int
	IsActive   *bool
	Limit      int
	Offset     int
}
