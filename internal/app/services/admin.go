package services

type AdminServices struct {
	RBAC      *RBACService
	Unit      *UnitService
	Room      *RoomService
	RoomType  *RoomTypeService
	MeterItem *MeterItemService
}
